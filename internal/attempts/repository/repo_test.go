package repository

import (
	"errors"
	"fmt"
	"strconv" // Import strconv
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

// TestAttemptRepository_SQLite là hàm test chính cho repository với SQLite
func TestAttemptRepository_SQLite(t *testing.T) {
	db := setupTestSQLiteDatabase(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewAttemptRepository(db)

	// --- Tạo dữ liệu phụ thuộc ---
	user1 := models.User{Name: "Student User 1", Email: "student1@test.com", Password: "pw", Role: "student", Status: "Active"}
	user2 := models.User{Name: "Student User 2", Email: "student2@test.com", Password: "pw", Role: "student", Status: "Active"}
	teacher := models.User{Name: "Teacher User", Email: "teacher@test.com", Password: "pw", Role: "teacher", Status: "Active"}
	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)
	require.NoError(t, db.Create(&teacher).Error)

	assessment1 := models.Assessment{Title: "Math Quiz", Subject: "Math", Duration: 30, CreatedByID: teacher.ID, Status: "active", PassingScore: 70, CreatedAt: time.Now().UTC().AddDate(0, 0, -1)}
	assessment2 := models.Assessment{Title: "Science Test", Subject: "Science", Duration: 60, CreatedByID: teacher.ID, Status: "active", PassingScore: 60, CreatedAt: time.Now().UTC().AddDate(0, 0, -1)}
	assessmentDraft := models.Assessment{Title: "Draft Quiz", Subject: "Draft", Duration: 10, CreatedByID: teacher.ID, Status: "draft", PassingScore: 50}
	require.NoError(t, db.Create(&assessment1).Error)
	require.NoError(t, db.Create(&assessment2).Error)
	require.NoError(t, db.Create(&assessmentDraft).Error)

	// Thêm settings cho assessment 1 để test FindAvailableAssessments
	setting1 := models.AssessmentSettings{AssessmentID: assessment1.ID, AllowRetake: true, MaxAttempts: 2}
	assessmentSettings2 := models.AssessmentSettings{
		AssessmentID:                2,
		RandomizeQuestions:          false,
		ShowResults:                 false,
		AllowRetake:                 false,
		MaxAttempts:                 1,
		TimeLimitEnforced:           false,
		RequireWebcam:               false,
		PreventTabSwitching:         false,
		RequireIdentityVerification: false,
	}
	require.NoError(t, db.Create(&setting1).Error)
	require.NoError(t, db.Create(&assessmentSettings2).Error)

	question1_1 := models.Question{AssessmentID: assessment1.ID, Type: "true-false", Text: "2+2=4?", CorrectAnswer: "true", Points: 5}
	question1_2 := models.Question{AssessmentID: assessment1.ID, Type: "essay", Text: "Describe Pi", Points: 10}
	require.NoError(t, db.Create(&question1_1).Error)
	require.NoError(t, db.Create(&question1_2).Error)

	// --- Biến lưu trữ ID/Đối tượng ---
	var createdAttemptID uint
	var attemptForAnswerID uint
	var answerToUpdateID uint
	// var attemptInProgressID uint // Sẽ gán giá trị sau khi tạo
	var inProgressAttempt *models.Attempt // Khai báo con trỏ ở đây

	// --- Sub-tests ---

	t.Run("TestCreateAttempt", func(t *testing.T) {
		now := time.Now()
		attempt := &models.Attempt{
			UserID:       user1.ID,
			AssessmentID: assessment1.ID,
			StartedAt:    now,
			Status:       "In Progress", // Tạo attempt này là In Progress ban đầu
		}
		err := repo.Create(attempt)
		assert.NoError(t, err)
		assert.NotZero(t, attempt.ID)
		createdAttemptID = attempt.ID
		attemptForAnswerID = attempt.ID // Dùng cho test answer

		// Kiểm tra trong DB
		var fetchedAttempt models.Attempt
		errFetch := db.First(&fetchedAttempt, createdAttemptID).Error
		assert.NoError(t, errFetch)
		assert.Equal(t, user1.ID, fetchedAttempt.UserID)
		assert.Equal(t, assessment1.ID, fetchedAttempt.AssessmentID)
		assert.Equal(t, "In Progress", fetchedAttempt.Status)
		assert.WithinDuration(t, now, fetchedAttempt.StartedAt, time.Second)
	})
	require.NotZero(t, createdAttemptID)

	// Tạo Attempt "In Progress" cho user2 để test IsUserInAttempt và ExpiredAttempt
	// Tạo sau TestCreate để đảm bảo createdAttemptID có giá trị
	inProgressAttempt = &models.Attempt{UserID: user2.ID, AssessmentID: assessment2.ID, StartedAt: time.Now(), Status: "In Progress"}
	require.NoError(t, repo.Create(inProgressAttempt), "Failed to create in-progress attempt for user 2")
	require.NotZero(t, inProgressAttempt.ID, "In-progress attempt ID should not be zero")

	t.Run("TestFindByID_Found", func(t *testing.T) {
		foundAttempt, err := repo.FindByID(createdAttemptID)
		assert.NoError(t, err)
		require.NotNil(t, foundAttempt)
		assert.Equal(t, createdAttemptID, foundAttempt.ID)
		assert.Equal(t, user1.ID, foundAttempt.UserID)
		assert.Equal(t, assessment1.ID, foundAttempt.AssessmentID)
	})

	t.Run("TestFindByID_NotFound", func(t *testing.T) {
		nonExistentID := uint(99999)
		foundAttempt, err := repo.FindByID(nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, foundAttempt)
		// Sửa lại kiểm tra lỗi cho chính xác hơn
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == fmt.Sprintf("attempt with ID %d not found", nonExistentID), "Expected ErrRecordNotFound or specific message")
	})

	t.Run("TestUpdateAttempt", func(t *testing.T) {
		attemptToUpdate, err := repo.FindByID(createdAttemptID)
		require.NoError(t, err)
		require.NotNil(t, attemptToUpdate)

		now := time.Now()
		submitted := now.Add(-1 * time.Minute)
		score := 85.5
		duration := 25
		attemptToUpdate.Status = "completed" // Cập nhật status
		attemptToUpdate.SubmittedAt = &submitted
		attemptToUpdate.EndedAt = &now
		attemptToUpdate.Score = &score
		attemptToUpdate.Duration = &duration
		attemptToUpdate.Feedback = "Good job"

		err = repo.Update(attemptToUpdate)
		assert.NoError(t, err)

		updatedAttempt, err := repo.FindByID(createdAttemptID)
		assert.NoError(t, err)
		require.NotNil(t, updatedAttempt)
		assert.Equal(t, "completed", updatedAttempt.Status) // Kiểm tra status đã cập nhật
		require.NotNil(t, updatedAttempt.SubmittedAt)
		assert.WithinDuration(t, submitted, *updatedAttempt.SubmittedAt, time.Second)
		require.NotNil(t, updatedAttempt.Score)
		assert.Equal(t, score, *updatedAttempt.Score)
		require.NotNil(t, updatedAttempt.Duration)
		assert.Equal(t, duration, *updatedAttempt.Duration)
		assert.Equal(t, "Good job", updatedAttempt.Feedback)
	})

	t.Run("TestSaveAnswer", func(t *testing.T) {
		answer := &models.Answer{
			AttemptID:  attemptForAnswerID,
			QuestionID: question1_1.ID,
			Answer:     "true",
			IsCorrect:  nil, // Chưa chấm
		}
		err := repo.SaveAnswer(answer)
		assert.NoError(t, err)
		assert.NotZero(t, answer.ID)
		answerToUpdateID = answer.ID // Lưu lại để test update

		// Kiểm tra trong DB
		var fetchedAnswer models.Answer
		errFetch := db.First(&fetchedAnswer, answer.ID).Error
		assert.NoError(t, errFetch)
		assert.Equal(t, attemptForAnswerID, fetchedAnswer.AttemptID)
		assert.Equal(t, question1_1.ID, fetchedAnswer.QuestionID)
		assert.Equal(t, "true", fetchedAnswer.Answer)
		assert.Nil(t, fetchedAnswer.IsCorrect)
	})
	require.NotZero(t, answerToUpdateID)

	t.Run("TestUpdateAnswer", func(t *testing.T) {
		answerToUpdate, err := repo.FindAnswerByAttemptAndQuestion(attemptForAnswerID, question1_1.ID)
		require.NoError(t, err)
		require.NotNil(t, answerToUpdate)
		require.Equal(t, answerToUpdateID, answerToUpdate.ID)

		isCorrect := true
		answerToUpdate.Answer = "false"       // Thay đổi câu trả lời
		answerToUpdate.IsCorrect = &isCorrect // Cập nhật trạng thái đúng/sai

		err = repo.UpdateAnswer(answerToUpdate)
		assert.NoError(t, err)

		// Kiểm tra lại
		updatedAnswer, err := repo.FindAnswerByAttemptAndQuestion(attemptForAnswerID, question1_1.ID)
		assert.NoError(t, err)
		require.NotNil(t, updatedAnswer)
		assert.Equal(t, "false", updatedAnswer.Answer)
		require.NotNil(t, updatedAnswer.IsCorrect)
		assert.True(t, *updatedAnswer.IsCorrect)
	})

	t.Run("TestFindAnswersByAttemptID", func(t *testing.T) {
		// Tạo thêm câu trả lời cho attemptForAnswerID
		answer2 := &models.Answer{AttemptID: attemptForAnswerID, QuestionID: question1_2.ID, Answer: "Essay answer"}
		require.NoError(t, repo.SaveAnswer(answer2))

		answers, err := repo.FindAnswersByAttemptID(attemptForAnswerID)
		assert.NoError(t, err)
		assert.Len(t, answers, 2) // Câu 1 đã update, câu 2 mới tạo

		foundQ1 := false
		foundQ2 := false
		for _, ans := range answers {
			if ans.QuestionID == question1_1.ID {
				foundQ1 = true
				assert.Equal(t, "false", ans.Answer) // Kiểm tra giá trị đã update
			}
			if ans.QuestionID == question1_2.ID {
				foundQ2 = true
				assert.Equal(t, "Essay answer", ans.Answer)
			}
		}
		assert.True(t, foundQ1)
		assert.True(t, foundQ2)
	})

	t.Run("TestFindAnswerByAttemptAndQuestion_Found", func(t *testing.T) {
		answer, err := repo.FindAnswerByAttemptAndQuestion(attemptForAnswerID, question1_1.ID)
		assert.NoError(t, err)
		require.NotNil(t, answer)
		assert.Equal(t, answerToUpdateID, answer.ID)
	})

	t.Run("TestFindAnswerByAttemptAndQuestion_NotFound", func(t *testing.T) {
		nonExistentQuestionID := uint(999)
		answer, err := repo.FindAnswerByAttemptAndQuestion(attemptForAnswerID, nonExistentQuestionID)
		assert.NoError(t, err) // GORM trả về nil, nil khi First không tìm thấy
		assert.Nil(t, answer)

		nonExistentAttemptID := uint(888)
		answer, err = repo.FindAnswerByAttemptAndQuestion(nonExistentAttemptID, question1_1.ID)
		assert.NoError(t, err)
		assert.Nil(t, answer)
	})

	t.Run("TestFindAvailableAssessments", func(t *testing.T) {
		// User 1 đã làm assessment 1 (status Completed)
		// User 1 chưa làm assessment 2 (Active)
		// User 1 chưa làm assessmentDraft (Draft)
		// User 1 có 1 attempt cho assessment 1, được phép 2 attempts (AllowRetake=true, MaxAttempts=2)
		var ass []models.Assessment
		err := db.Model(&models.Assessment{}).Find(&ass).Error
		assert.NoError(t, err)
		params := util.PaginationParams{Page: 0, Limit: 10}
		available, total, err := repo.FindAvailableAssessments(user1.ID, params)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2)) // Phải thấy assessment 1 (còn lượt) và assessment 2
		assert.LessOrEqual(t, len(available), 10)

		foundAssessment1 := false
		foundAssessment2 := false
		foundDraft := false

		for _, item := range available {
			id_str, ok_id := item["id"].(string) // SQLite có thể trả về string
			if !ok_id {
				if id_int, ok_int := item["id"].(int64); ok_int { // Hoặc int64
					id_str = strconv.FormatInt(id_int, 10)
				} else {
					t.Logf("Could not assert type for id: %v", item["id"])
					continue // Bỏ qua nếu không lấy được ID
				}
			}
			id, _ := strconv.ParseUint(id_str, 10, 64)

			title, _ := item["title"].(string)
			// SQLite có thể trả về tên cột khác nhau hoặc kiểu khác nhau
			allowRetakeVal, _ := item["allow_retake"]
			maxAttemptsVal, _ := item["max_attempts"]
			attemptCountVal, _ := item["attempt_count"]
			canAttemptVal, _ := item["can_attempt"]

			allowRetake := false
			if allowRetakeInt, ok := allowRetakeVal.(int64); ok { // SQLite bool có thể là int 0/1
				allowRetake = allowRetakeInt == 1
			} else if allowRetakeBool, ok := allowRetakeVal.(bool); ok {
				allowRetake = allowRetakeBool
			}

			canAttempt := false
			if canAttemptInt, ok := canAttemptVal.(int64); ok {
				canAttempt = canAttemptInt == 1
			} else if canAttemptBool, ok := canAttemptVal.(bool); ok {
				canAttempt = canAttemptBool
			}

			if uint(id) == assessment1.ID {
				foundAssessment1 = true
				assert.Equal(t, "Math Quiz", title)
				assert.True(t, allowRetake)
				assert.Equal(t, (2), maxAttemptsVal)
				assert.Equal(t, (1), attemptCountVal) // User 1 đã làm 1 lần
				assert.True(t, canAttempt)            // Vẫn còn lượt
			}
			if uint(id) == assessment2.ID {
				foundAssessment2 = true
				assert.Equal(t, "Science Test", title)
			}
			if title == "Draft Quiz" { // Không nên thấy bài Draft
				foundDraft = true
			}
		}

		assert.True(t, foundAssessment1, "Assessment 1 should be available (retake allowed)")
		assert.True(t, foundAssessment2, "Assessment 2 should be available")
		assert.False(t, foundDraft, "Draft assessment should not be available")
	})

	t.Run("TestHasCompletedAssessment", func(t *testing.T) {
		// User 1 đã hoàn thành assessment 1
		completed, err := repo.HasCompletedAssessment(user1.ID, assessment1.ID)
		assert.NoError(t, err)
		assert.True(t, completed)

		// User 1 chưa hoàn thành assessment 2
		completed, err = repo.HasCompletedAssessment(user1.ID, assessment2.ID)
		assert.NoError(t, err)
		assert.False(t, completed)

		// User 2 chưa hoàn thành assessment 1
		completed, err = repo.HasCompletedAssessment(user2.ID, assessment1.ID)
		assert.NoError(t, err)
		assert.False(t, completed)
	})

	t.Run("TestCountAttemptsByUserAndAssessment", func(t *testing.T) {
		// User 1 có 1 attempt cho assessment 1
		count, err := repo.CountAttemptsByUserAndAssessment(user1.ID, assessment1.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		// User 2 có 1 attempt cho assessment 2 (là inProgressAttempt)
		count, err = repo.CountAttemptsByUserAndAssessment(user2.ID, assessment2.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		// User 2 không có attempt nào cho assessment 1
		count, err = repo.CountAttemptsByUserAndAssessment(user2.ID, assessment1.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("TestFindCompletedAttemptsByUserAndAssessment", func(t *testing.T) {
		// User 1 có 1 attempt đã completed cho assessment 1
		results, err := repo.FindCompletedAttemptsByUserAndAssessment(user1.ID, assessment1.ID)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, float64(85.5), results[0]["score"]) // Điểm đã update
		assert.Equal(t, "completed", results[0]["status"])

		// User 2 không có attempt completed nào cho assessment 1
		results, err = repo.FindCompletedAttemptsByUserAndAssessment(user2.ID, assessment1.ID)
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("TestGetAllAttemptByUserId", func(t *testing.T) {
		// User 1 có 1 attempt (assessment 1 - Completed)
		params := util.PaginationParams{Page: 0, Limit: 10}
		attempts, total, err := repo.GetAllAttemptByUserId(user1.ID, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, attempts, 1)
		assert.Equal(t, createdAttemptID, attempts[0].ID)

		// User 2 có 1 attempt (assessment 2 - In Progress)
		attempts, total, err = repo.GetAllAttemptByUserId(user2.ID, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, attempts, 1)
		assert.Equal(t, inProgressAttempt.ID, attempts[0].ID)
	})

	t.Run("TestListAttemptByUserAndAssessmentID", func(t *testing.T) {
		params := util.PaginationParams{Page: 0, Limit: 10}
		// User 1, Assessment 1
		attempts, total, err := repo.ListAttemptByUserAndAssessmentID(user1.ID, assessment1.ID, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, attempts, 1)
		assert.Equal(t, createdAttemptID, attempts[0].ID)

		// User 1, Assessment 2 (chưa làm)
		attempts, total, err = repo.ListAttemptByUserAndAssessmentID(user1.ID, assessment2.ID, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, attempts, 0)

		// User 2, Assessment 2 (đang làm)
		attempts, total, err = repo.ListAttemptByUserAndAssessmentID(user2.ID, assessment2.ID, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, attempts, 1)
		assert.Equal(t, inProgressAttempt.ID, attempts[0].ID)
	})

	// Bỏ qua các test thống kê phức tạp cho SQLite
	// t.Run("TestGetAssessmentCompletionRates", ...)
	// t.Run("TestGetScoreDistribution", ...)
	// t.Run("TestGetAverageTimeSpent", ...)
	// t.Run("TestGetMostChallengingAssessments", ...)
	// t.Run("TestGetMostSuccessfulAssessments", ...)
	// t.Run("TestGetPassRate", ...)
	// t.Run("TestCountAll", ...)
	// t.Run("TestCountByPeriod", ...)

	t.Run("TestSaveSuspiciousActivity", func(t *testing.T) {
		activity := &models.SuspiciousActivity{
			UserID:       user1.ID,
			AssessmentID: assessment1.ID,
			AttemptID:    createdAttemptID,
			Type:         "TAB_SWITCH",
			Details:      "User switched tabs",
			Timestamp:    time.Now(),
			Severity:     "HIGH",
		}
		err := repo.SaveSuspiciousActivity(activity)
		assert.NoError(t, err)
		assert.NotZero(t, activity.ID)

		// Kiểm tra DB
		var fetchedActivity models.SuspiciousActivity
		errFetch := db.First(&fetchedActivity, activity.ID).Error
		assert.NoError(t, errFetch)
		assert.Equal(t, activity.Type, fetchedActivity.Type)
		assert.Equal(t, activity.AttemptID, fetchedActivity.AttemptID)
	})

	t.Run("TestFindSuspiciousActivitiesByAttemptID", func(t *testing.T) {
		// Tạo thêm activity
		activity2 := &models.SuspiciousActivity{AttemptID: createdAttemptID, UserID: user1.ID, AssessmentID: assessment1.ID, Type: "LOOKING_AWAY"}
		require.NoError(t, repo.SaveSuspiciousActivity(activity2))

		activities, err := repo.FindSuspiciousActivitiesByAttemptID(createdAttemptID)
		assert.NoError(t, err)
		assert.Len(t, activities, 2)
	})

	t.Run("TestIsUserInAttempt", func(t *testing.T) {
		// User 1 có attempt đã completed
		inAttempt, err := repo.IsUserInAttempt(user1.ID)
		assert.NoError(t, err)
		assert.False(t, inAttempt)

		// User 2 có attempt đang diễn ra
		inAttempt, err = repo.IsUserInAttempt(user2.ID)
		assert.NoError(t, err)
		assert.True(t, inAttempt)

	})

	t.Run("TestExpiredAttempt", func(t *testing.T) {
		// Sử dụng attempt đang diễn ra của user2 đã tạo ở test trước
		expired, err := repo.ExpiredAttempt() // Hàm này chỉ lấy các attempt "In Progress"
		assert.NoError(t, err)
		require.NotEmpty(t, expired, "Should find the in-progress attempt") // Phải tìm thấy attempt của user2

		foundUser2Attempt := false
		for _, att := range expired {
			if att.ID == inProgressAttempt.ID { // Kiểm tra đúng ID của attempt đang chạy
				foundUser2Attempt = true
				assert.Equal(t, "In Progress", att.Status)
				break
			}
		}
		assert.True(t, foundUser2Attempt, "In-progress attempt for user 2 not found in expired check")

		// Kiểm tra không chứa attempt đã completed của user1
		foundUser1Attempt := false
		for _, att := range expired {
			if att.ID == createdAttemptID {
				foundUser1Attempt = true
				break
			}
		}
		assert.False(t, foundUser1Attempt, "Completed attempt should not be in expired check")
	})

	t.Run("TestDeleteAttempt", func(t *testing.T) {
		// Sử dụng attempt đã completed của user1
		err := repo.Delete(createdAttemptID)
		assert.NoError(t, err)

		// Thử tìm lại
		_, err = repo.FindByID(createdAttemptID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == fmt.Sprintf("attempt with ID %d not found", createdAttemptID))

		// Kiểm tra soft delete (nếu cần)
		var deletedAttempt models.Attempt
		errUnscoped := db.Unscoped().First(&deletedAttempt, createdAttemptID).Error
		assert.NoError(t, errUnscoped) // Tìm thấy nếu dùng Unscoped
		assert.NotNil(t, deletedAttempt.DeletedAt)
		assert.True(t, deletedAttempt.DeletedAt.Valid) // Kiểm tra cờ Valid của gorm.DeletedAt
	})

}
