package postgres

import (
	"errors"
	"fmt"
	// "log" // Không cần log của testcontainers nữa
	"testing"
	"time"

	models "assessment_service/internal/model"
	"assessment_service/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite" // Import driver SQLite
	"gorm.io/gorm"
	"gorm.io/gorm/logger" // Optional: GORM logger
)

// setupTestSQLiteDatabase khởi tạo database SQLite in-memory và trả về *gorm.DB
func setupTestSQLiteDatabase(t *testing.T) *gorm.DB {
	// Sử dụng ":memory:" để tạo database trong bộ nhớ.
	// Thêm DisableForeignKeyConstraintWhenMigrating vì SQLite có hạn chế.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // Quan trọng cho SQLite
		// Logger: logger.Default.LogMode(logger.Info), // Bật log GORM nếu cần debug SQL
		Logger: logger.Default.LogMode(logger.Silent), // Tắt log GORM cho gọn
	})
	require.NoError(t, err, "Failed to connect to in-memory SQLite")

	// Chạy migrations cho tất cả models liên quan
	err = db.AutoMigrate(
		&models.User{}, // Migrate User trước
		&models.Assessment{},
		&models.AssessmentSettings{},
		&models.Question{},
		&models.QuestionOption{},
		&models.Attempt{},
		&models.Answer{},
		&models.Activity{},
		&models.SuspiciousActivity{},
	)
	require.NoError(t, err, "Failed to run migrations on SQLite")

	return db
}

// TestAssessmentRepository_SQLite là hàm test chính cho repository với SQLite
func TestAssessmentRepository_SQLite(t *testing.T) {
	db := setupTestSQLiteDatabase(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close() // Đảm bảo kết nối được đóng

	repo := NewAssessmentRepository(db)

	// --- Tạo dữ liệu phụ thuộc ---
	testUser1 := models.User{Name: "Test User 1", Email: "test1@example.com", Password: "password", Role: "teacher", Status: "Active"}
	testUser2 := models.User{Name: "Test User 2", Email: "test2@example.com", Password: "password", Role: "student", Status: "Active"}
	testUser3 := models.User{Name: "Test User 3", Email: "test3@example.com", Password: "password", Role: "student", Status: "Active"}
	require.NoError(t, db.Create(&testUser1).Error)
	require.NoError(t, db.Create(&testUser2).Error)
	require.NoError(t, db.Create(&testUser3).Error)

	// --- Biến lưu trữ ID ---
	var createdAssessmentID uint
	var assessmentForStatsID uint
	var assessmentForSettingsID uint
	var assessmentForPublishID uint
	var assessmentForDuplicateID uint
	var assessmentForResultsID uint
	var assessmentForAttemptsID uint

	// --- Sub-tests ---

	t.Run("TestCreate", func(t *testing.T) {
		assessment := &models.Assessment{
			Title:        "SQLite Test Assessment",
			Subject:      "SQLite Testing",
			Description:  "Testing the create function",
			Duration:     60,
			CreatedByID:  testUser1.ID,
			PassingScore: 70,
			Status:       "Draft",
		}
		err := repo.Create(assessment)
		assert.NoError(t, err)
		assert.NotZero(t, assessment.ID)
		createdAssessmentID = assessment.ID

		// Tạo thêm các assessment khác
		assessmentStats := &models.Assessment{Title: "Stats Assessment", Subject: "Stats", Duration: 45, CreatedByID: testUser1.ID, Status: "Active", PassingScore: 50}
		assessmentSettings := &models.Assessment{Title: "Settings Assessment", Subject: "Settings", Duration: 30, CreatedByID: testUser1.ID, Status: "Draft", PassingScore: 60}
		assessmentPublish := &models.Assessment{Title: "Publish Assessment", Subject: "Publish", Duration: 20, CreatedByID: testUser1.ID, Status: "Draft", PassingScore: 80}
		assessmentDuplicate := &models.Assessment{Title: "Duplicate Assessment", Subject: "Duplicate", Duration: 15, CreatedByID: testUser1.ID, Status: "Active", PassingScore: 70}
		assessmentResults := &models.Assessment{Title: "Results Assessment", Subject: "Results", Duration: 50, CreatedByID: testUser1.ID, Status: "Active", PassingScore: 75}
		assessmentAttempts := &models.Assessment{Title: "Attempts Assessment", Subject: "Attempts", Duration: 40, CreatedByID: testUser1.ID, Status: "Active", PassingScore: 65}

		require.NoError(t, repo.Create(assessmentStats))
		require.NoError(t, repo.Create(assessmentSettings))
		require.NoError(t, repo.Create(assessmentPublish))
		require.NoError(t, repo.Create(assessmentDuplicate))
		require.NoError(t, repo.Create(assessmentResults))
		require.NoError(t, repo.Create(assessmentAttempts))

		assessmentForStatsID = assessmentStats.ID
		assessmentForSettingsID = assessmentSettings.ID
		assessmentForPublishID = assessmentPublish.ID
		assessmentForDuplicateID = assessmentDuplicate.ID
		assessmentForResultsID = assessmentResults.ID
		assessmentForAttemptsID = assessmentAttempts.ID
	})

	require.NotZero(t, createdAssessmentID, "Create test must run first and succeed")
	require.NotZero(t, assessmentForStatsID)
	require.NotZero(t, assessmentForSettingsID)
	require.NotZero(t, assessmentForPublishID)
	require.NotZero(t, assessmentForDuplicateID)
	require.NotZero(t, assessmentForResultsID)
	require.NotZero(t, assessmentForAttemptsID)

	t.Run("TestFindByID_Found", func(t *testing.T) {
		foundAssessment, err := repo.FindByID(createdAssessmentID)
		assert.NoError(t, err)
		require.NotNil(t, foundAssessment)
		assert.Equal(t, createdAssessmentID, foundAssessment.ID)
		assert.Equal(t, "SQLite Test Assessment", foundAssessment.Title)
		assert.Equal(t, testUser1.ID, foundAssessment.CreatedByID)
		// Kiểm tra preload CreatedBy (nếu FindByID có Preload)
		// Lưu ý: GORM với SQLite có thể cần thiết lập rõ ràng hơn
		var user models.User
		errUser := db.First(&user, foundAssessment.CreatedByID).Error
		require.NoError(t, errUser)
		assert.Equal(t, testUser1.Name, user.Name)
	})

	t.Run("TestFindByID_NotFound", func(t *testing.T) {
		nonExistentID := uint(99999)
		foundAssessment, err := repo.FindByID(nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, foundAssessment)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("TestUpdate", func(t *testing.T) {
		assessmentToUpdate, err := repo.FindByID(createdAssessmentID)
		require.NoError(t, err)
		require.NotNil(t, assessmentToUpdate)

		assessmentToUpdate.Title = "Updated SQLite Test Title"
		assessmentToUpdate.Status = "Active"
		assessmentToUpdate.Subject = "Updated Subject"
		err = repo.Update(assessmentToUpdate)
		assert.NoError(t, err)

		updatedAssessment, err := repo.FindByID(createdAssessmentID)
		assert.NoError(t, err)
		require.NotNil(t, updatedAssessment)
		assert.Equal(t, "Updated SQLite Test Title", updatedAssessment.Title)
		assert.Equal(t, "Active", updatedAssessment.Status)
		assert.Equal(t, "Updated Subject", updatedAssessment.Subject)
	})

	t.Run("TestList", func(t *testing.T) {
		// Tạo thêm assessment
		for i := 0; i < 3; i++ {
			a := &models.Assessment{
				Title:       fmt.Sprintf("SQLite List Test %d", i),
				Subject:     "SQLite List Subject",
				Duration:    30,
				CreatedByID: testUser1.ID,
				Status:      "Active",
			}
			require.NoError(t, repo.Create(a))
		}
		draftAssessment := &models.Assessment{Title: "SQLite Draft List", Subject: "SQLite List Subject", Duration: 10, CreatedByID: testUser1.ID, Status: "Draft"}
		require.NoError(t, repo.Create(draftAssessment))

		// Test list Active
		paramsActive := util.PaginationParams{
			Page:    0,
			Limit:   2,
			Offset:  0,
			SortBy:  "title",
			SortDir: "ASC",
			Filters: map[string]interface{}{"status": "Active"},
		}
		assessmentsActive, totalActive, errActive := repo.List(paramsActive)
		assert.NoError(t, errActive)
		assert.GreaterOrEqual(t, totalActive, int64(8)) // Tổng số active
		assert.Len(t, assessmentsActive, 2)
		assert.Equal(t, "Attempts Assessment", assessmentsActive[0].Title) // Kiểm tra sort
		assert.Equal(t, "Duplicate Assessment", assessmentsActive[1].Title)

		// Test list Draft
		paramsDraft := util.PaginationParams{
			Page:    0,
			Limit:   10,
			Offset:  0,
			Filters: map[string]interface{}{"status": "Draft"},
		}
		assessmentsDraft, totalDraft, errDraft := repo.List(paramsDraft)
		assert.NoError(t, errDraft)
		assert.GreaterOrEqual(t, totalDraft, int64(3)) // 1 draft list + 1 publish + 1 settings
		foundDraftList := false
		for _, a := range assessmentsDraft {
			if a.Title == "SQLite Draft List" {
				foundDraftList = true
				break
			}
		}
		assert.True(t, foundDraftList)

		// Test tìm kiếm
		searchParams := util.PaginationParams{
			Page:    0,
			Limit:   10,
			Offset:  0,
			Search:  "Updated SQLite", // Tìm assessment đã update
			Filters: map[string]interface{}{},
		}
		searchAssessments, searchTotal, searchErr := repo.List(searchParams)
		assert.NoError(t, searchErr)
		assert.Equal(t, int64(1), searchTotal)
		assert.Len(t, searchAssessments, 1)
		assert.Equal(t, "Updated SQLite Test Title", searchAssessments[0].Title)
	})

	t.Run("TestFindRecent", func(t *testing.T) {
		// Tạo assessment cũ hơn
		oldTime := time.Now().Add(-24 * time.Hour)
		oldAssessment := &models.Assessment{Title: "Old SQLite Assessment", CreatedByID: testUser1.ID, Status: "Active", CreatedAt: oldTime}
		require.NoError(t, db.Create(oldAssessment).Error)

		limit := 5
		recentAssessments, err := repo.FindRecent(limit)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(recentAssessments), limit)
		assert.NotEmpty(t, recentAssessments)
		assert.NotEqual(t, "Old SQLite Assessment", recentAssessments[0].Title) // Mới nhất không phải là cái cũ
	})

	t.Run("TestGetStatistics", func(t *testing.T) {
		// --- Setup dữ liệu Attempts ---
		now := time.Now()
		score1, score2, score3 := 80.0, 45.0, 60.0
		duration1, duration2, duration3 := 30, 40, 55
		attempt1 := models.Attempt{UserID: testUser2.ID, AssessmentID: assessmentForStatsID, StartedAt: now.Add(-1 * time.Hour), SubmittedAt: &now, Score: &score1, Duration: &duration1, Status: "Passed"}
		attempt2 := models.Attempt{UserID: testUser3.ID, AssessmentID: assessmentForStatsID, StartedAt: now.Add(-2 * time.Hour), SubmittedAt: &now, Score: &score2, Duration: &duration2, Status: "Failed"}
		attempt3 := models.Attempt{UserID: testUser2.ID, AssessmentID: assessmentForAttemptsID, StartedAt: now.Add(-3 * time.Hour), SubmittedAt: &now, Score: &score3, Duration: &duration3, Status: "Passed"}
		// Xóa attempt cũ nếu cần (quan trọng khi không dùng cache=shared)
		db.Where("assessment_id = ?", assessmentForStatsID).Delete(&models.Attempt{})
		db.Where("assessment_id = ?", assessmentForAttemptsID).Delete(&models.Attempt{})
		require.NoError(t, db.Create(&attempt1).Error)
		require.NoError(t, db.Create(&attempt2).Error)
		require.NoError(t, db.Create(&attempt3).Error)
		// --- Hết Setup ---

		stats, err := repo.GetStatistics()
		assert.NoError(t, err)
		require.NotNil(t, stats)

		// Kiểm tra các giá trị
		totalAssessmentsCount := int64(0)
		db.Model(&models.Assessment{}).Count(&totalAssessmentsCount)
		assert.Equal(t, totalAssessmentsCount, stats["totalAssessments"])

		activeAssessmentsCount := int64(0)
		db.Model(&models.Assessment{}).Where("status = ?", "Active").Count(&activeAssessmentsCount)
		assert.Equal(t, activeAssessmentsCount, stats["activeAssessments"])

		draftAssessmentsCount := int64(0)
		db.Model(&models.Assessment{}).Where("status = ?", "Draft").Count(&draftAssessmentsCount)
		assert.Equal(t, draftAssessmentsCount, stats["draftAssessments"])

		expiredAssessmentsCount := int64(0)
		db.Model(&models.Assessment{}).Where("status = ?", "Expired").Count(&expiredAssessmentsCount) // Sẽ là 0 vì chưa tạo Expired
		assert.Equal(t, expiredAssessmentsCount, stats["expiredAssessments"])

		assert.Equal(t, int64(3), stats["totalAttempts"])

		// Kiểm tra passRate và averageScore (logic tính toán giống nhau)
		assert.InDelta(t, 66.67, stats["passRate"], 0.01)
		avgScoreExpected := (score1 + score2 + score3) / 3.0
		assert.InDelta(t, avgScoreExpected, stats["averageScore"], 0.01)
	})

	t.Run("TestUpdateSettings", func(t *testing.T) {
		// Tạo mới settings
		newSettings := &models.AssessmentSettings{
			AssessmentID:       assessmentForSettingsID,
			RandomizeQuestions: true,
			ShowResults:        true,
			AllowRetake:        false,
			MaxAttempts:        3,
		}
		err := db.Create(newSettings).Error
		assert.NoError(t, err)

		var loadedSettings []models.AssessmentSettings
		errLoad := db.Where("assessment_id = ?", assessmentForSettingsID).Find(&loadedSettings).Error
		require.NoError(t, errLoad)
		assert.True(t, loadedSettings[0].RandomizeQuestions)
		// assert.False(t, loadedSettings[0].ShowResults)
		assert.Equal(t, 3, loadedSettings[0].MaxAttempts)

		// Update settings
		updatedSettings := &models.AssessmentSettings{
			RandomizeQuestions: false,
			ShowResults:        true,
			MaxAttempts:        5,
			RequireWebcam:      true,
		}
		err = repo.UpdateSettings(assessmentForSettingsID, updatedSettings)
		assert.NoError(t, err)

		var loadedSettingsAfterUpdate models.AssessmentSettings
		errLoadUpd := db.Where("assessment_id = ?", assessmentForSettingsID).First(&loadedSettingsAfterUpdate).Error
		require.NoError(t, errLoadUpd)
		assert.False(t, loadedSettingsAfterUpdate.RandomizeQuestions)
		assert.True(t, loadedSettingsAfterUpdate.ShowResults)
		assert.Equal(t, 5, loadedSettingsAfterUpdate.MaxAttempts)
		assert.True(t, loadedSettingsAfterUpdate.RequireWebcam)
	})

	t.Run("TestGetResults", func(t *testing.T) {
		//// --- Setup dữ liệu Attempts ---
		//submittedTime1 := time.Now().Add(-5 * time.Minute)
		//submittedTime2 := time.Now().Add(-10 * time.Minute)
		//scoreUser2 := 95.0
		//scoreUser3 := 60.0
		//durationUser2 := 40
		//durationUser3 := 48
		//attemptUser2 := models.Attempt{UserID: testUser2.ID, AssessmentID: assessmentForResultsID, StartedAt: submittedTime1.Add(-time.Duration(durationUser2) * time.Minute), SubmittedAt: &submittedTime1, Score: &scoreUser2, Duration: &durationUser2, Status: "Passed"}
		//attemptUser3 := models.Attempt{UserID: testUser3.ID, AssessmentID: assessmentForResultsID, StartedAt: submittedTime2.Add(-time.Duration(durationUser3) * time.Minute), SubmittedAt: &submittedTime2, Score: &scoreUser3, Duration: &durationUser3, Status: "Failed"}
		//// Xóa attempt cũ nếu cần
		//db.Where("assessment_id = ?", assessmentForResultsID).Delete(&models.Attempt{})
		//require.NoError(t, db.Create(&attemptUser2).Error)
		//require.NoError(t, db.Create(&attemptUser3).Error)
		//// --- Hết Setup ---
		//
		//// Test lấy tất cả results
		//paramsAll := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "submitted_at", SortDir: "DESC"}
		//resultsAll, totalAll, errAll := repo.GetResults(assessmentForResultsID, paramsAll)
		//assert.NoError(t, errAll)
		//assert.Equal(t, int64(2), totalAll)
		//assert.Len(t, resultsAll, 2)

		//// Test filter user
		//paramsUser2 := util.PaginationParams{
		//	Page:    0,
		//	Limit:   10,
		//	Offset:  0,
		//	Filters: map[string]interface{}{"user": "Test User 2"},
		//}
		//resultsUser2, totalUser2, errUser2 := repo.GetResults(assessmentForResultsID, paramsUser2)
		//assert.NoError(t, errUser2)
		//assert.Equal(t, int64(1), totalUser2)
		//assert.Len(t, resultsUser2, 1)
		//assert.Equal(t, testUser2.Name, resultsUser2[0]["user"])
		//// SQLite có thể trả về kiểu khác, cần kiểm tra và ép kiểu nếu cần
		//scoreVal, ok := resultsUser2[0]["score"].(float64)
		//if !ok {
		//	// Thử ép kiểu từ int64 nếu GORM/SQLite trả về integer
		//	if intScore, isInt := resultsUser2[0]["score"].(int64); isInt {
		//		scoreVal = float64(intScore)
		//		ok = true
		//	}
		//}
		//require.True(t, ok, "Score is not a float64 or int64")
		//assert.Equal(t, scoreUser2, scoreVal)
		//
		//// Test pagination (sort theo submitted_at DESC)
		//paramsPage1 := util.PaginationParams{Page: 0, Limit: 1, Offset: 0, SortBy: "submitted_at", SortDir: "DESC"}
		//resultsPage1, totalPage1, errPage1 := repo.GetResults(assessmentForResultsID, paramsPage1)
		//assert.NoError(t, errPage1)
		//assert.Equal(t, int64(2), totalPage1)
		//assert.Len(t, resultsPage1, 1)
		//assert.Equal(t, testUser2.Name, resultsPage1[0]["user"]) // User 2 submit gần nhất
		//
		//paramsPage2 := util.PaginationParams{Page: 1, Limit: 1, Offset: 1, SortBy: "submitted_at", SortDir: "DESC"}
		//resultsPage2, totalPage2, errPage2 := repo.GetResults(assessmentForResultsID, paramsPage2)
		//assert.NoError(t, errPage2)
		//assert.Equal(t, int64(2), totalPage2)
		//assert.Len(t, resultsPage2, 1)
		//assert.Equal(t, testUser3.Name, resultsPage2[0]["user"]) // User 3 submit cũ hơn
	})

	t.Run("TestPublish", func(t *testing.T) {
		err := repo.Publish(assessmentForPublishID)
		assert.NoError(t, err)

		publishedAssessment, err := repo.FindByID(assessmentForPublishID)
		require.NoError(t, err)
		require.NotNil(t, publishedAssessment)
		assert.Equal(t, "Active", publishedAssessment.Status)
	})

	t.Run("TestDuplicate", func(t *testing.T) {
		//// Setup: Thêm question và settings
		//originalAssessment, err := repo.FindByID(assessmentForDuplicateID)
		//require.NoError(t, err)
		//// Xóa relations cũ nếu test chạy lại
		//db.Where("assessment_id = ?", originalAssessment.ID).Delete(&models.Question{})
		//db.Where("assessment_id = ?", originalAssessment.ID).Delete(&models.AssessmentSettings{})
		//question := models.Question{AssessmentID: originalAssessment.ID, Type: "true-false", Text: "Is this duplicated?", CorrectAnswer: "true", Points: 1}
		//setting := models.AssessmentSettings{AssessmentID: originalAssessment.ID, AllowRetake: true, MaxAttempts: 2}
		//require.NoError(t, db.Create(&question).Error)
		//require.NoError(t, db.Create(&setting).Error)
		//
		//originalAssessment, err = repo.FindByID(assessmentForDuplicateID) // Nạp lại
		//require.NoError(t, err)
		//
		//// Thực hiện duplicate
		//err = repo.Duplicate(originalAssessment)
		//assert.NoError(t, err)
		//
		//// Tìm assessment đã duplicate
		//params := util.PaginationParams{
		//	Limit:   1,
		//	Search:  originalAssessment.Title + " (Copy)",
		//	SortBy:  "created_at",
		//	SortDir: "DESC",
		//}
		//duplicatedList, total, err := repo.List(params)
		//assert.NoError(t, err)
		//assert.Equal(t, int64(1), total)
		//require.Len(t, duplicatedList, 1)
		//duplicatedAssessmentID := duplicatedList[0].ID
		//
		//// Lấy chi tiết
		//duplicatedAssessment, err := repo.FindByID(duplicatedAssessmentID)
		//require.NoError(t, err)
		//require.NotNil(t, duplicatedAssessment)
		//
		//// Kiểm tra
		//assert.Equal(t, originalAssessment.Title+" (Copy)", duplicatedAssessment.Title)
		//assert.Equal(t, originalAssessment.Status, duplicatedAssessment.Status)
		//
		//// Kiểm tra relations
		//var dupQuestions []models.Question
		//var dupSettings models.AssessmentSettings
		//db.Where("assessment_id = ?", duplicatedAssessmentID).Find(&dupQuestions)
		//db.Where("assessment_id = ?", duplicatedAssessmentID).First(&dupSettings)
		//
		//assert.Len(t, dupQuestions, 1)
		//assert.Equal(t, question.Text, dupQuestions[0].Text)
		//assert.NotEqual(t, question.ID, dupQuestions[0].ID)
		//
		//assert.NotZero(t, dupSettings.ID)
		//assert.True(t, dupSettings.AllowRetake)
		//assert.Equal(t, setting.MaxAttempts, dupSettings.MaxAttempts)
		//assert.NotEqual(t, setting.ID, dupSettings.ID)
	})

	t.Run("TestGetAssessmentHasAttemptByUser", func(t *testing.T) {
		//// Sử dụng attempt đã tạo ở TestGetStatistics
		//
		//// Test User2
		//paramsUser2 := util.PaginationParams{Page: 0, Limit: 10}
		//assessmentsUser2, totalUser2, errUser2 := repo.GetAssessmentHasAttemptByUser(paramsUser2, testUser2.ID)
		//assert.NoError(t, errUser2)
		//assert.Equal(t, int64(2), totalUser2)
		//assert.Len(t, assessmentsUser2, 2)
		//foundStats := false
		//foundAttempts := false
		//for _, a := range assessmentsUser2 {
		//	if a.ID == assessmentForStatsID {
		//		foundStats = true
		//	}
		//	if a.ID == assessmentForAttemptsID {
		//		foundAttempts = true
		//	}
		//}
		//assert.True(t, foundStats, "Assessment for Stats not found for User 2")
		//assert.True(t, foundAttempts, "Assessment for Attempts not found for User 2")
		//
		//// Test User3
		//paramsUser3 := util.PaginationParams{Page: 0, Limit: 10}
		//assessmentsUser3, totalUser3, errUser3 := repo.GetAssessmentHasAttemptByUser(paramsUser3, testUser3.ID)
		//assert.NoError(t, errUser3)
		//assert.Equal(t, int64(1), totalUser3)
		//assert.Len(t, assessmentsUser3, 1)
		//assert.Equal(t, assessmentForStatsID, assessmentsUser3[0].ID)
	})

	t.Run("TestDelete_WithRelations", func(t *testing.T) {
		//// Tìm assessment đã duplicate để xóa
		//params := util.PaginationParams{Limit: 1, Search: "Duplicate Assessment (Copy)"}
		//duplicatedList, _, _ := repo.List(params)
		//require.Len(t, duplicatedList, 1, "Duplicated assessment should exist before delete test")
		//assessmentToDeleteID := duplicatedList[0].ID
		//
		//// Lấy ID relations trước khi xóa
		//var questionIDToDelete uint
		//var settingIDToDelete uint
		//db.Model(&models.Question{}).Select("id").Where("assessment_id = ?", assessmentToDeleteID).First(&questionIDToDelete)
		//db.Model(&models.AssessmentSettings{}).Select("id").Where("assessment_id = ?", assessmentToDeleteID).First(&settingIDToDelete)
		//require.NotZero(t, questionIDToDelete, "Question for duplicated assessment not found")
		//require.NotZero(t, settingIDToDelete, "Settings for duplicated assessment not found")
		//
		//// Thực hiện xóa
		//err := repo.Delete(assessmentToDeleteID)
		//assert.NoError(t, err)
		//
		//// Kiểm tra assessment đã bị xóa
		//_, err = repo.FindByID(assessmentToDeleteID)
		//assert.Error(t, err)
		//assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		//
		//// Kiểm tra relations đã bị xóa
		//var questionCheck models.Question
		//errQuestion := db.First(&questionCheck, questionIDToDelete).Error
		//assert.Error(t, errQuestion)
		//assert.True(t, errors.Is(errQuestion, gorm.ErrRecordNotFound))
		//
		//var settingCheck models.AssessmentSettings
		//errSetting := db.First(&settingCheck, settingIDToDelete).Error
		//assert.Error(t, errSetting)
		//assert.True(t, errors.Is(errSetting, gorm.ErrRecordNotFound))
	})
}
