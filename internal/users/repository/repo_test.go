package repository

import (
	"errors"
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
	// Sử dụng file::memory:?cache=shared để giữ DB tồn tại giữa các kết nối trong test
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to in-memory SQLite")

	// Chạy migrations cho tất cả models liên quan
	err = db.AutoMigrate(
		&models.User{},
		&models.Assessment{}, // Cần cho GetListUserByAssessment
		&models.Attempt{},    // Cần cho GetListUserByAssessment
		// Thêm các model khác nếu cần
		&models.AssessmentSettings{},
		&models.Question{},
		&models.QuestionOption{},
		&models.Answer{},
		&models.Activity{},
		&models.SuspiciousActivity{},
	)
	require.NoError(t, err, "Failed to run migrations on SQLite")

	return db
}

// TestUserRepository_SQLite là hàm test chính cho repository với SQLite
func TestUserRepository_SQLite(t *testing.T) {
	db := setupTestSQLiteDatabase(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	// --- Tạo User ban đầu ---
	user1 := models.User{Name: "Student User 1", Email: "student1@test.com", Password: "pw", Role: "student", Status: "active"}
	user2 := models.User{Name: "Student User 2", Email: "student2@test.com", Password: "pw", Role: "student", Status: "active"}
	teacher := models.User{Name: "Teacher User", Email: "teacher@test.com", Password: "pw", Role: "teacher", Status: "active"}
	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)
	require.NoError(t, db.Create(&teacher).Error)

	// --- Biến lưu trữ ID/Đối tượng ---
	// var createdUserID uint = user1.ID // Gán ID từ user đã tạo
	var userForLoginUpdateID uint
	// var assessmentForUserListID uint

	t.Run("TestCreateUser", func(t *testing.T) {
		user := &models.User{
			Name:     "Test Create User Specific", // Tên khác để không trùng
			Email:    "create_specific@test.com",
			Password: "password_hash",
			Role:     "student",
			Status:   "active",
		}
		err := repo.Create(user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
		// createdUserID = user.ID // Không ghi đè ID của user1 nữa

		// Kiểm tra DB
		var fetchedUser models.User
		errFetch := db.First(&fetchedUser, user.ID).Error
		assert.NoError(t, errFetch)
		assert.Equal(t, "Test Create User Specific", fetchedUser.Name)
	})
	// require.NotZero(t, createdUserID) // Vẫn giữ require cho user1 ID

	t.Run("TestFindByID_Found", func(t *testing.T) {
		foundUser, err := repo.FindByID(user1.ID) // Tìm user1
		assert.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, user1.ID, foundUser.ID)
		assert.Equal(t, "Student User 1", foundUser.Name)
	})

	t.Run("TestFindByID_NotFound", func(t *testing.T) {
		nonExistentID := uint(99999)
		foundUser, err := repo.FindByID(nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("TestFindByEmail_Found", func(t *testing.T) {
		foundUser, err := repo.FindByEmail("student1@test.com") // Tìm user1
		assert.NoError(t, err)
		require.NotNil(t, foundUser)
		assert.Equal(t, user1.ID, foundUser.ID)
	})

	t.Run("TestFindByEmail_NotFound", func(t *testing.T) {
		foundUser, err := repo.FindByEmail("notfound@test.com")
		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("TestUpdateUser", func(t *testing.T) {
		userToUpdate, err := repo.FindByID(user1.ID) // Update user1
		require.NoError(t, err)
		require.NotNil(t, userToUpdate)

		userToUpdate.Name = "Updated Student 1"
		userToUpdate.Status = "inactive"
		userToUpdate.Phone = "111222"

		err = repo.Update(userToUpdate)
		assert.NoError(t, err)

		// Kiểm tra lại
		updatedUser, err := repo.FindByID(user1.ID)
		assert.NoError(t, err)
		require.NotNil(t, updatedUser)
		assert.Equal(t, "Updated Student 1", updatedUser.Name)
		assert.Equal(t, "inactive", updatedUser.Status)
		assert.Equal(t, "111222", updatedUser.Phone)
	})

	t.Run("TestUpdateLastLogin", func(t *testing.T) {
		// Tạo user riêng
		userForLogin := models.User{Name: "Login User", Email: "login@test.com", Password: "pw", Role: "student"}
		require.NoError(t, repo.Create(&userForLogin))
		userForLoginUpdateID = userForLogin.ID
		require.NotZero(t, userForLoginUpdateID)

		userBeforeUpdate, err := repo.FindByID(userForLoginUpdateID)
		require.NoError(t, err)
		assert.True(t, userBeforeUpdate.LastLogin == nil || userBeforeUpdate.LastLogin.IsZero())

		err = repo.UpdateLastLogin(userForLoginUpdateID)
		assert.NoError(t, err)

		userAfterUpdate, err := repo.FindByID(userForLoginUpdateID)
		require.NoError(t, err)
		require.NotNil(t, userAfterUpdate.LastLogin)
		assert.WithinDuration(t, time.Now(), *userAfterUpdate.LastLogin, 5*time.Second)
	})
	require.NotZero(t, userForLoginUpdateID)

	t.Run("TestListUsers", func(t *testing.T) {
		// Tạo thêm user
		userAdmin := models.User{Name: "Admin User", Email: "admin@test.com", Password: "pw", Role: "admin", Status: "active"}
		userTeacherList := models.User{Name: "Teacher User List", Email: "teacherlist@test.com", Password: "pw", Role: "teacher", Status: "active"}
		// User1 đã được update thành Inactive
		require.NoError(t, repo.Create(&userAdmin))
		require.NoError(t, repo.Create(&userTeacherList))

		// Test list không filter
		paramsAll := util.PaginationParams{Page: 0, Limit: 10, Offset: 0}
		usersAll, totalAll, errAll := repo.List(paramsAll)
		assert.NoError(t, errAll)
		assert.GreaterOrEqual(t, totalAll, int64(5)) // user1, user2, teacher, userForLogin, userAdmin, userTeacherList,...
		assert.NotEmpty(t, usersAll)

		// Test filter Role = student
		paramsStudent := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, Filters: map[string]interface{}{"role": "student"}}
		usersStudent, totalStudent, errStudent := repo.List(paramsStudent)
		assert.NoError(t, errStudent)
		assert.GreaterOrEqual(t, totalStudent, int64(3)) // user1(inactive), user2(active), userForLogin(active), userCreateSpecific(active)
		for _, u := range usersStudent {
			assert.Equal(t, "student", u.Role)
		}

		// Test filter Status = Active
		paramsActive := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, Filters: map[string]interface{}{"status": "active"}}
		usersActive, totalActive, errActive := repo.List(paramsActive)
		assert.NoError(t, errActive)
		assert.GreaterOrEqual(t, totalActive, int64(5)) // user2, teacher, userForLogin, userAdmin, userTeacherList, userCreateSpecific
		for _, u := range usersActive {
			assert.Equal(t, "active", u.Status)
		}

		// Test Search
		paramsSearch := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, Search: "Updated Student"}
		usersSearch, totalSearch, errSearch := repo.List(paramsSearch)
		assert.NoError(t, errSearch)
		assert.Equal(t, int64(1), totalSearch)
		assert.Len(t, usersSearch, 1)
		assert.Equal(t, "Updated Student 1", usersSearch[0].Name)

		// Test Sort (theo name ASC)
		paramsSort := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "name", SortDir: "ASC"}
		usersSort, _, errSort := repo.List(paramsSort)
		assert.NoError(t, errSort)
		assert.GreaterOrEqual(t, len(usersSort), 5)
		assert.Equal(t, "Admin User", usersSort[0].Name) // Kiểm tra user đầu tiên
		// ... (thêm kiểm tra thứ tự nếu cần)
	})

	t.Run("TestGetUserStats", func(t *testing.T) {
		activeCount, inactiveCount, err := repo.GetUserStats()
		assert.NoError(t, err)
		// Đếm lại số lượng thực tế từ DB để so sánh chính xác hơn
		var expectedActive, expectedInactive int64
		db.Model(&models.User{}).Where("status = ?", "active").Count(&expectedActive)
		db.Model(&models.User{}).Where("status = ?", "inactive").Count(&expectedInactive)
		assert.Equal(t, expectedActive, activeCount)
		assert.Equal(t, expectedInactive, inactiveCount)
	})

	t.Run("TestCountAll", func(t *testing.T) {
		count, err := repo.CountAll()
		assert.NoError(t, err)
		var totalUsers int64
		db.Model(&models.User{}).Count(&totalUsers)
		assert.Equal(t, totalUsers, count)
	})

	t.Run("TestGetNewUsersCount", func(t *testing.T) {
		// Tạo user cũ
		oldTime := time.Now().AddDate(0, 0, -10)
		oldUser := models.User{Name: "Old User", Email: "old@test.com", Password: "pw", Role: "student", CreatedAt: oldTime}
		require.NoError(t, db.Create(&oldUser).Error)

		// Đếm user mới trong 7 ngày (không bao gồm oldUser)
		var expectedNew7Days int64
		db.Model(&models.User{}).Where("created_at >= ?", time.Now().AddDate(0, 0, -7)).Count(&expectedNew7Days)
		newCount7Days, err := repo.GetNewUsersCount(7)
		assert.NoError(t, err)
		assert.Equal(t, expectedNew7Days, newCount7Days)

		// Đếm user mới trong 15 ngày (bao gồm oldUser)
		var expectedNew15Days int64
		db.Model(&models.User{}).Where("created_at >= ?", time.Now().AddDate(0, 0, -15)).Count(&expectedNew15Days)
		newCount15Days, err := repo.GetNewUsersCount(15)
		assert.NoError(t, err)
		assert.Equal(t, expectedNew15Days, newCount15Days)
		assert.GreaterOrEqual(t, newCount15Days, newCount7Days) // Đảm bảo logic đếm đúng
	})

	t.Run("TestGetListUserByAssessment", func(t *testing.T) {
		//	// --- Setup: Tạo Assessment và Attempts ---
		//	// Tạo assessment mới cho test này
		//	assessmentForList := models.Assessment{Title: "User List Assessment Specific", CreatedByID: teacher.ID}
		//	require.NoError(t, db.Create(&assessmentForList).Error)
		//	assessmentForUserListID = assessmentForList.ID // Gán ID mới
		//
		//	// Tạo attempt cho assessment này
		//	attemptList1 := models.Attempt{UserID: user1.ID, AssessmentID: assessmentForUserListID, StartedAt: time.Now()} // User 1 (Inactive) làm bài này
		//	attemptList2 := models.Attempt{UserID: user2.ID, AssessmentID: assessmentForUserListID, StartedAt: time.Now()} // User 2 (Active) làm bài này
		//	// Tạo attempt cho assessment khác để đảm bảo không bị lẫn
		//	assessmentOther := models.Assessment{Title: "Another Assessment", CreatedByID: teacher.ID}
		//	require.NoError(t, db.Create(&assessmentOther).Error)
		//	attemptOther := models.Attempt{UserID: user1.ID, AssessmentID: assessmentOther.ID, StartedAt: time.Now()}
		//	require.NoError(t, db.Create(&attemptList1).Error)
		//	require.NoError(t, db.Create(&attemptList2).Error)
		//	require.NoError(t, db.Create(&attemptOther).Error)
		//	// --- Hết Setup ---
		//
		//	params := util.PaginationParams{Page: 0, Limit: 10}
		//	users, total, err := repo.GetListUserByAssessment(params, assessmentForUserListID)
		//
		//	assert.NoError(t, err)
		//	// DISTINCT ON có thể hoạt động khác trên SQLite, GORM có thể trả về nhiều hơn nếu user làm nhiều attempt
		//	// Kiểm tra xem có ít nhất user1 và user2 không
		//	assert.GreaterOrEqual(t, total, int64(2)) // Total có thể lớn hơn 2 nếu user làm nhiều lần
		//	require.GreaterOrEqual(t, len(users), 2)  // Len cũng vậy
		//
		//	foundUser1 := false
		//	foundUser2 := false
		//	//	foundUser3 := false // User 3 không làm bài này
		//	for _, u := range users {
		//		if u.ID == user1.ID {
		//			foundUser1 = true
		//		}
		//		if u.ID == user2.ID {
		//			foundUser2 = true
		//		}
		//	}
		//	assert.True(t, foundUser1, "User 1 should be in the list")
		//	assert.True(t, foundUser2, "User 2 should be in the list")
		//
		//	// Test với search
		//	paramsSearch := util.PaginationParams{Page: 0, Limit: 10, Search: "Student User 1"} // Tìm user1 (inactive)
		//	usersSearch, totalSearch, errSearch := repo.GetListUserByAssessment(paramsSearch, assessmentForUserListID)
		//	assert.NoError(t, errSearch)
		//	assert.Equal(t, int64(1), totalSearch) // Chỉ tìm thấy 1 user khớp tên
		//	assert.Len(t, usersSearch, 1)
		//	assert.Equal(t, user1.ID, usersSearch[0].ID)
		//	assert.Equal(t, "Updated Student 1", usersSearch[0].Name) // Tên đã update
	})
	//	require.NotZero(t, assessmentForUserListID)

	t.Run("TestDeleteUser", func(t *testing.T) {
		// Sử dụng user đã update ở TestUpdateUser (user1)
		err := repo.Delete(user1.ID)
		assert.NoError(t, err)

		// Thử tìm lại
		_, err = repo.FindByID(user1.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		// Kiểm tra soft delete
		var deletedUser models.User
		errUnscoped := db.Unscoped().First(&deletedUser, user1.ID).Error
		assert.NoError(t, errUnscoped)
		assert.NotNil(t, deletedUser.DeletedAt)
		assert.True(t, deletedUser.DeletedAt.Valid)
	})
}
