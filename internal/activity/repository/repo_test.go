package repository

import (
	"testing"
	"time"

	models "assessment_service/internal/model"
	"assessment_service/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestSQLiteDatabase khởi tạo database SQLite in-memory và trả về *gorm.DB
func setupTestSQLiteDatabase(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to in-memory SQLite")

	// Chạy migrations cho tất cả models liên quan
	err = db.AutoMigrate(
		&models.User{},
		&models.Assessment{},
		&models.Attempt{}, // Activity có thể liên quan đến Attempt
		&models.Activity{},
		&models.SuspiciousActivity{},
	)
	require.NoError(t, err, "Failed to run migrations on SQLite")

	return db
}

// TestActivityRepository_SQLite là hàm test chính cho repository với SQLite
func TestActivityRepository_SQLite(t *testing.T) {
	db := setupTestSQLiteDatabase(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewActivityRepository(db)

	// --- Tạo dữ liệu phụ thuộc ---
	user1 := models.User{Name: "Activity User 1", Email: "activity1@test.com", Password: "pw", Role: "student"}
	user2 := models.User{Name: "Activity User 2", Email: "activity2@test.com", Password: "pw", Role: "student"}
	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)

	assessment1 := models.Assessment{Title: "Activity Assessment", CreatedByID: user1.ID /*Hoặc ID của teacher nếu có*/}
	require.NoError(t, db.Create(&assessment1).Error)

	attempt1User1 := models.Attempt{UserID: user1.ID, AssessmentID: assessment1.ID, StartedAt: time.Now()}
	require.NoError(t, db.Create(&attempt1User1).Error)

	// --- Biến lưu trữ ID/Đối tượng ---
	var createdActivityID uint
	//var createdSuspiciousActivityID uint

	t.Run("TestCreateActivity", func(t *testing.T) {
		now := time.Now()
		activity := &models.Activity{
			UserID:       user1.ID,
			Action:       "LOGIN",
			Details:      "User logged in",
			IPAddress:    "127.0.0.1",
			UserAgent:    "test-agent",
			Timestamp:    now,
			AssessmentID: &assessment1.ID,
		}
		err := repo.Create(activity)
		assert.NoError(t, err)
		assert.NotZero(t, activity.ID)
		createdActivityID = activity.ID

		var fetchedActivity models.Activity
		errFetch := db.First(&fetchedActivity, createdActivityID).Error
		assert.NoError(t, errFetch)
		assert.Equal(t, user1.ID, fetchedActivity.UserID)
		assert.Equal(t, "LOGIN", fetchedActivity.Action)
		assert.Equal(t, assessment1.ID, *fetchedActivity.AssessmentID)
		assert.WithinDuration(t, now, fetchedActivity.Timestamp, time.Second)
	})
	require.NotZero(t, createdActivityID)

	t.Run("TestFindByUserID", func(t *testing.T) {
		// Tạo thêm activity cho user1
		activity2 := &models.Activity{UserID: user1.ID, Action: "VIEW_ASSESSMENT", Timestamp: time.Now().Add(-1 * time.Hour)}
		require.NoError(t, repo.Create(activity2))
		// Tạo activity cho user2
		activity3 := &models.Activity{UserID: user2.ID, Action: "START_ASSESSMENT", Timestamp: time.Now()}
		require.NoError(t, repo.Create(activity3))

		params := util.PaginationParams{Page: 0, Limit: 10}
		activitiesUser1, totalUser1, err := repo.FindByUserID(user1.ID, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), totalUser1)
		assert.Len(t, activitiesUser1, 2)
		// Kiểm tra sắp xếp mặc định (timestamp DESC)
		assert.True(t, activitiesUser1[0].Timestamp.After(activitiesUser1[1].Timestamp))

		// Test filter theo thời gian
		paramsFiltered := util.PaginationParams{
			Page:  0,
			Limit: 10,
			Filters: map[string]interface{}{
				"from": time.Now().Add(-30 * time.Minute).Format("2006-01-02 15:04:05"), // Filter activity trong 30 phút gần nhất
			},
		}
		activitiesFiltered, totalFiltered, errFiltered := repo.FindByUserID(user1.ID, paramsFiltered)
		assert.NoError(t, errFiltered)
		assert.Equal(t, int64(1), totalFiltered) // Chỉ có activity "LOGIN"
		assert.Len(t, activitiesFiltered, 1)
		assert.Equal(t, "LOGIN", activitiesFiltered[0].Action)
	})

	t.Run("TestBulkCreateActivities", func(t *testing.T) {
		assessmentID := assessment1.ID // Gán giá trị cho assessmentID
		activities := []models.Activity{
			{UserID: user1.ID, Action: "BULK_ACTION_1", Timestamp: time.Now(), AssessmentID: &assessmentID},
			{UserID: user1.ID, Action: "BULK_ACTION_2", Timestamp: time.Now().Add(1 * time.Second), AssessmentID: &assessmentID},
		}
		err := repo.BulkCreate(activities)
		assert.NoError(t, err)

		// Kiểm tra xem các activity đã được tạo chưa
		var count int64
		db.Model(&models.Activity{}).Where("user_id = ? AND action LIKE ?", user1.ID, "BULK_ACTION_%").Count(&count)
		assert.Equal(t, int64(2), count)
	})

	t.Run("TestFindByAssessmentID", func(t *testing.T) {
		// Các activity đã tạo ở trên đều thuộc assessment1.ID
		params := util.PaginationParams{Page: 0, Limit: 10}
		activities, total, err := repo.FindByAssessmentID(assessment1.ID, params)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3)) // LOGIN, BULK_ACTION_1, BULK_ACTION_2
		assert.NotEmpty(t, activities)
		for _, act := range activities {
			require.NotNil(t, act.AssessmentID)
			assert.Equal(t, assessment1.ID, *act.AssessmentID)
		}
	})

	t.Run("TestSaveAndFindSuspiciousActivity", func(t *testing.T) {
		//suspicious := &models.SuspiciousActivity{
		//	UserID:       user1.ID,
		//	AssessmentID: assessment1.ID,
		//	AttemptID:    attempt1User1.ID,
		//	Type:         "TAB_SWITCH",
		//	Details:      "User switched tabs multiple times",
		//	Timestamp:    time.Now(),
		//	Severity:     "HIGH",
		//}
		//err := repo.SaveSuspiciousActivity(suspicious) // Giả sử hàm này có trong repo interface
		//assert.NoError(t, err)
		//assert.NotZero(t, suspicious.ID)
		//createdSuspiciousActivityID = suspicious.ID
		//
		//// Test FindSuspiciousActivity
		//params := util.PaginationParams{Page: 0, Limit: 10}
		//activities, total, err := repo.FindSuspiciousActivity(user1.ID, attempt1User1.ID, params)
		//assert.NoError(t, err)
		//assert.Equal(t, int64(1), total)
		//assert.Len(t, activities, 1)
		//assert.Equal(t, createdSuspiciousActivityID, activities[0].ID)
		//assert.Equal(t, "TAB_SWITCH", activities[0].Type)
	})
	// 	require.NotZero(t, createdSuspiciousActivityID)

	t.Run("TestCountByPeriod", func(t *testing.T) {
		//// Tạo 1 activity cũ
		//oldTime := time.Now().AddDate(0, 0, -2)
		//oldActivity := models.Activity{UserID: user1.ID, Action: "OLD_ACTION", Timestamp: oldTime}
		//require.NoError(t, repo.Create(&oldActivity))
		//
		//// Đếm activity trong 1 ngày gần nhất (bao gồm các activity đã tạo ở các test trước)
		//countLastDay, err := repo.CountByPeriod(1)
		//assert.NoError(t, err)
		//assert.GreaterOrEqual(t, countLastDay, int64(5)) // LOGIN, VIEW_ASSESSMENT, BULK_1, BULK_2, Suspicious (nếu được tính)
		//
		//// Đếm activity trong 3 ngày gần nhất (bao gồm cả oldActivity)
		//countLast3Days, err := repo.CountByPeriod(3)
		//assert.NoError(t, err)
		//assert.GreaterOrEqual(t, countLast3Days, countLastDay+1)
	})

	// --- Các hàm thống kê (chỉ kiểm tra chạy không lỗi trên SQLite) ---
	// Lưu ý: Các hàm này sử dụng Raw SQL và có thể cần điều chỉnh cho SQLite
	// trong file repo.go để hoạt động đúng hoặc trả về kết quả có ý nghĩa.

	t.Run("TestGetDailyActiveUsers_SQLite", func(t *testing.T) {
		// Tạo thêm dữ liệu activity cho vài ngày
		db.Create(&models.Activity{UserID: user1.ID, Action: "ACTION_D1", Timestamp: time.Now()})
		db.Create(&models.Activity{UserID: user2.ID, Action: "ACTION_D1", Timestamp: time.Now()})
		db.Create(&models.Activity{UserID: user1.ID, Action: "ACTION_D2", Timestamp: time.Now().AddDate(0, 0, -1)})

		result, err := repo.GetDailyActiveUsers(3)
		assert.NoError(t, err) // Quan trọng nhất là không lỗi cú pháp
		assert.NotNil(t, result)
		// Kiểm tra cấu trúc cơ bản nếu có kết quả
		if len(result) > 0 {
			_, dateOk := result[0]["date"] // SQLite có thể trả về date dạng string
			_, countOk := result[0]["count"]
			assert.True(t, dateOk, "Expected 'date' key in daily active users result")
			assert.True(t, countOk, "Expected 'count' key in daily active users result")
		}
	})

	t.Run("TestGetActivityByHour_SQLite", func(t *testing.T) {
		//result, err := repo.GetActivityByHour()
		//assert.NoError(t, err)
		//assert.NotNil(t, result)
		//if len(result) > 0 {
		//	_, hourOk := result[0]["hour"]
		//	_, countOk := result[0]["count"]
		//	assert.True(t, hourOk)
		//	assert.True(t, countOk)
		//}
	})

	t.Run("TestGetActivityByType_SQLite", func(t *testing.T) {
		//result, err := repo.GetActivityByType()
		//assert.NoError(t, err)
		//assert.NotNil(t, result)
		//if len(result) > 0 {
		//	_, typeOk := result[0]["type"] // Hoặc "action" tùy vào alias trong SQL
		//	_, countOk := result[0]["count"]
		//	assert.True(t, typeOk)
		//	assert.True(t, countOk)
		//}
	})

	t.Run("TestGetTotalActiveUsers_SQLite", func(t *testing.T) {
		//count, err := repo.GetTotalActiveUsers()
		//assert.NoError(t, err)
		//assert.GreaterOrEqual(t, count, int64(0)) // Chỉ cần không lỗi và trả về số >= 0
	})

	t.Run("TestGetRecentActivity_SQLite", func(t *testing.T) {
		// Tạo activity gần đây
		db.Create(&models.Activity{UserID: user1.ID, User: user1, Action: "RECENT_ACTION", Timestamp: time.Now().Add(-5 * time.Minute), AssessmentID: &assessment1.ID})
		result, err := repo.GetRecentActivity(1) // 1 giờ gần nhất
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if len(result) > 0 {
			item := result[0]
			_, idOk := item["id"]
			_, typeOk := item["type"]
			_, userOk := item["user"]
			assert.True(t, idOk)
			assert.True(t, typeOk)
			assert.True(t, userOk)
		}
	})

	t.Run("TestGetActiveUsers_SQLite", func(t *testing.T) {
		//count, err := repo.GetActiveUsers(15) // 15 phút gần nhất
		//assert.NoError(t, err)
		//assert.GreaterOrEqual(t, count, int64(0))
	})

	t.Run("TestGetTrending_SQLite", func(t *testing.T) {
		//// Cần Assessment và Activity liên kết
		//result, err := repo.GetTrending()
		//assert.NoError(t, err)
		//assert.NotNil(t, result)
		//if len(result) > 0 {
		//	item := result[0]
		//	_, idOk := item["id"]
		//	_, titleOk := item["title"]
		//	_, activityCountOk := item["activity_count"]
		//	assert.True(t, idOk)
		//	assert.True(t, titleOk)
		//	assert.True(t, activityCountOk)
		//}
	})
}
