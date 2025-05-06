package repository

import (
	models "assessment_service/internal/model"
	"errors"
	"testing"

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

	// Chạy migrations
	err = db.AutoMigrate(
		&models.User{}, // Cần cho Assessment.CreatedByID
		&models.Assessment{},
		&models.Question{},
		&models.QuestionOption{},
	)
	require.NoError(t, err, "Failed to run migrations on SQLite")

	return db
}

// TestQuestionRepository_SQLite là hàm test chính cho repository với SQLite
func TestQuestionRepository_SQLite(t *testing.T) {
	db := setupTestSQLiteDatabase(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	repo := NewQuestionRepository(db)

	// --- Tạo dữ liệu phụ thuộc ---
	testUser := models.User{Name: "Question Tester", Email: "q@test.com", Password: "pw", Role: "teacher"}
	require.NoError(t, db.Create(&testUser).Error)

	assessment1 := models.Assessment{Title: "Assessment For Questions 1", CreatedByID: testUser.ID, Duration: 30}
	assessment2 := models.Assessment{Title: "Assessment For Questions 2", CreatedByID: testUser.ID, Duration: 60}
	require.NoError(t, db.Create(&assessment1).Error)
	require.NoError(t, db.Create(&assessment2).Error)

	// --- Biến lưu trữ ID/Đối tượng ---
	var createdQuestionID uint
	var createdOptionID uint

	t.Run("TestCreateQuestion_MC", func(t *testing.T) {
		question := &models.Question{
			AssessmentID:  assessment1.ID,
			Type:          "multiple-choice",
			Text:          "What is 1+1?",
			CorrectAnswer: "b",
			Points:        10,
			Options: []models.QuestionOption{
				{OptionID: "a", Text: "1"},
				{OptionID: "b", Text: "2"},
				{OptionID: "c", Text: "3"},
			},
		}
		err := repo.Create(question)
		assert.NoError(t, err)
		assert.NotZero(t, question.ID)
		createdQuestionID = question.ID

		// Kiểm tra DB
		var fetchedQuestion models.Question
		errFetch := db.Preload("Options").First(&fetchedQuestion, createdQuestionID).Error
		assert.NoError(t, errFetch)
		assert.Equal(t, assessment1.ID, fetchedQuestion.AssessmentID)
		assert.Equal(t, "multiple-choice", fetchedQuestion.Type)
		assert.Equal(t, "What is 1+1?", fetchedQuestion.Text)
		assert.Equal(t, "b", fetchedQuestion.CorrectAnswer)
		assert.Len(t, fetchedQuestion.Options, 3)
		assert.Equal(t, "a", fetchedQuestion.Options[0].OptionID) // GORM thường sort theo PK
		assert.Equal(t, "b", fetchedQuestion.Options[1].OptionID)
		assert.Equal(t, "c", fetchedQuestion.Options[2].OptionID)
	})
	require.NotZero(t, createdQuestionID)

	t.Run("TestCreateQuestion_Essay", func(t *testing.T) {
		question := &models.Question{
			AssessmentID: assessment1.ID,
			Type:         "essay",
			Text:         "Describe testing.",
			Points:       20,
			// Không có Options và CorrectAnswer
		}
		err := repo.Create(question)
		assert.NoError(t, err)
		assert.NotZero(t, question.ID)

		var fetchedQuestion models.Question
		errFetch := db.Preload("Options").First(&fetchedQuestion, question.ID).Error
		assert.NoError(t, errFetch)
		assert.Equal(t, "essay", fetchedQuestion.Type)
		assert.Empty(t, fetchedQuestion.CorrectAnswer)
		assert.Empty(t, fetchedQuestion.Options) // Không có options được tạo
	})

	t.Run("TestFindByID_Found", func(t *testing.T) {
		foundQuestion, err := repo.FindByID(createdQuestionID)
		assert.NoError(t, err)
		require.NotNil(t, foundQuestion)
		assert.Equal(t, createdQuestionID, foundQuestion.ID)
		assert.Equal(t, "What is 1+1?", foundQuestion.Text)
		assert.Len(t, foundQuestion.Options, 3) // Kiểm tra options được preload
	})

	t.Run("TestFindByID_NotFound", func(t *testing.T) {
		nonExistentID := uint(99999)
		foundQuestion, err := repo.FindByID(nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, foundQuestion)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("TestFindByAssessmentID", func(t *testing.T) {
		// assessment1 có 2 questions đã tạo
		questions, err := repo.FindByAssessmentID(assessment1.ID)
		assert.NoError(t, err)
		assert.Len(t, questions, 2)

		// assessment2 chưa có question nào
		questions, err = repo.FindByAssessmentID(assessment2.ID)
		assert.NoError(t, err)
		assert.Len(t, questions, 0)
	})

	t.Run("TestUpdateQuestion", func(t *testing.T) {
		questionToUpdate, err := repo.FindByID(createdQuestionID)
		require.NoError(t, err)
		require.NotNil(t, questionToUpdate)

		questionToUpdate.Text = "Updated: What is 2+2?"
		questionToUpdate.CorrectAnswer = "c" // Giả sử đáp án đúng thay đổi
		questionToUpdate.Points = 15

		// Lưu ý: Hàm Update của repo này không tự động cập nhật Options.
		// Việc cập nhật Options thường được xử lý riêng biệt (thêm/xóa/sửa option).
		err = repo.Update(questionToUpdate)
		assert.NoError(t, err)

		// Kiểm tra lại
		updatedQuestion, err := repo.FindByID(createdQuestionID)
		assert.NoError(t, err)
		require.NotNil(t, updatedQuestion)
		assert.Equal(t, "Updated: What is 2+2?", updatedQuestion.Text)
		assert.Equal(t, "c", updatedQuestion.CorrectAnswer)
		assert.Equal(t, 15.0, updatedQuestion.Points) // GORM có thể đọc float
	})

	t.Run("TestAddOption", func(t *testing.T) {
		newOption := &models.QuestionOption{
			QuestionID: createdQuestionID,
			OptionID:   "d",
			Text:       "4",
		}
		err := repo.AddOption(newOption)
		assert.NoError(t, err)
		assert.NotZero(t, newOption.ID)
		createdOptionID = newOption.ID // Lưu ID để test update/delete

		// Kiểm tra lại question xem có option mới không
		questionWithOptions, err := repo.FindByID(createdQuestionID)
		assert.NoError(t, err)
		require.NotNil(t, questionWithOptions)
		assert.Len(t, questionWithOptions.Options, 4) // Giờ phải là 4 options
		foundNew := false
		for _, opt := range questionWithOptions.Options {
			if opt.OptionID == "d" {
				foundNew = true
				assert.Equal(t, "4", opt.Text)
				break
			}
		}
		assert.True(t, foundNew)
	})
	require.NotZero(t, createdOptionID)

	t.Run("TestUpdateOption", func(t *testing.T) {
		// Lấy option vừa tạo để update
		var optionToUpdate models.QuestionOption
		err := db.First(&optionToUpdate, createdOptionID).Error
		require.NoError(t, err)

		optionToUpdate.Text = "Four" // Update text
		err = repo.UpdateOption(&optionToUpdate)
		assert.NoError(t, err)

		// Kiểm tra lại
		var updatedOption models.QuestionOption
		err = db.First(&updatedOption, createdOptionID).Error
		assert.NoError(t, err)
		assert.Equal(t, "Four", updatedOption.Text)
		assert.Equal(t, "d", updatedOption.OptionID) // OptionID không đổi
	})

	t.Run("TestDeleteOption", func(t *testing.T) {
		err := repo.DeleteOption(createdOptionID)
		assert.NoError(t, err)

		// Kiểm tra xem option còn tồn tại không
		var deletedOption models.QuestionOption
		err = db.First(&deletedOption, createdOptionID).Error
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		// Kiểm tra lại question xem số option đã giảm chưa
		questionAfterDelete, err := repo.FindByID(createdQuestionID)
		assert.NoError(t, err)
		require.NotNil(t, questionAfterDelete)
		assert.Len(t, questionAfterDelete.Options, 3) // Quay về 3 options
	})

	t.Run("TestDeleteQuestion", func(t *testing.T) {
		// Lấy question và các options liên quan trước khi xóa
		questionToDelete, err := repo.FindByID(createdQuestionID)
		require.NoError(t, err)
		require.NotNil(t, questionToDelete)
		require.Len(t, questionToDelete.Options, 3)
		optionIDsToDelete := []uint{}
		for _, opt := range questionToDelete.Options {
			optionIDsToDelete = append(optionIDsToDelete, opt.ID)
		}

		// Thực hiện xóa question
		err = repo.Delete(createdQuestionID)
		assert.NoError(t, err)

		// Kiểm tra question đã bị xóa
		_, err = repo.FindByID(createdQuestionID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		// Kiểm tra các options liên quan đã bị xóa (do transaction trong repo.Delete)
		var remainingOptions int64
		db.Model(&models.QuestionOption{}).Where("id IN ?", optionIDsToDelete).Count(&remainingOptions)
		assert.Zero(t, remainingOptions)
	})

}
