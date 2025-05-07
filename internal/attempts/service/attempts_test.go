package service

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
	"testing"
)

// --- Mock AttemptRepository ---
type MockAttemptRepository struct {
	mock.Mock
}

// Implement AttemptRepository interface for mock
func (m *MockAttemptRepository) Create(attempt *models.Attempt) error {
	args := m.Called(attempt)
	return args.Error(0)
}

func (m *MockAttemptRepository) FindByID(id uint) (*models.Attempt, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	attempt, ok := args.Get(0).(*models.Attempt)
	if !ok && args.Get(0) != nil {
		panic("Mock FindByID returned non-nil value of incorrect type")
	}
	return attempt, args.Error(1)
}

func (m *MockAttemptRepository) Update(attempt *models.Attempt) error {
	args := m.Called(attempt)
	return args.Error(0)
}

func (m *MockAttemptRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAttemptRepository) SaveAnswer(answer *models.Answer) error {
	args := m.Called(answer)
	return args.Error(0)
}

func (m *MockAttemptRepository) UpdateAnswer(answer *models.Answer) error {
	args := m.Called(answer)
	return args.Error(0)
}

func (m *MockAttemptRepository) FindAnswersByAttemptID(attemptID uint) ([]models.Answer, error) {
	args := m.Called(attemptID)
	answers, _ := args.Get(0).([]models.Answer)
	return answers, args.Error(1)
}

func (m *MockAttemptRepository) FindAnswerByAttemptAndQuestion(attemptID, questionID uint) (*models.Answer, error) {
	args := m.Called(attemptID, questionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	answer, ok := args.Get(0).(*models.Answer)
	if !ok && args.Get(0) != nil {
		panic("Mock FindAnswerByAttemptAndQuestion returned non-nil value of incorrect type")
	}
	return answer, args.Error(1)
}

func (m *MockAttemptRepository) FindAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(userID, params)
	results, _ := args.Get(0).([]map[string]interface{})
	count, _ := args.Get(1).(int64)
	return results, count, args.Error(2)
}

func (m *MockAttemptRepository) HasCompletedAssessment(userID, assessmentID uint) (bool, error) {
	args := m.Called(userID, assessmentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAttemptRepository) CountAttemptsByUserAndAssessment(userID, assessmentID uint) (int, error) {
	args := m.Called(userID, assessmentID)
	return args.Int(0), args.Error(1)
}

func (m *MockAttemptRepository) FindCompletedAttemptsByUserAndAssessment(userID, assessmentID uint) ([]map[string]interface{}, error) {
	args := m.Called(userID, assessmentID)
	results, _ := args.Get(0).([]map[string]interface{})
	return results, args.Error(1)
}

func (m *MockAttemptRepository) GetAllAttemptByUserId(userID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, params)
	attempts, _ := args.Get(0).([]models.Attempt)
	count, _ := args.Get(1).(int64)
	return attempts, count, args.Error(2)
}

func (m *MockAttemptRepository) ListAttemptByUserAndAssessmentID(userID uint, assessmentID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, assessmentID, params)
	attempts, _ := args.Get(0).([]models.Attempt)
	count, _ := args.Get(1).(int64)
	return attempts, count, args.Error(2)
}

func (m *MockAttemptRepository) GetAssessmentCompletionRates() (map[string]interface{}, error) {
	args := m.Called()
	stats, _ := args.Get(0).(map[string]interface{})
	return stats, args.Error(1)
}

func (m *MockAttemptRepository) GetScoreDistribution() (map[string]interface{}, error) {
	args := m.Called()
	stats, _ := args.Get(0).(map[string]interface{})
	return stats, args.Error(1)
}

func (m *MockAttemptRepository) GetAverageTimeSpent() (map[string]interface{}, error) {
	args := m.Called()
	stats, _ := args.Get(0).(map[string]interface{})
	return stats, args.Error(1)
}

func (m *MockAttemptRepository) GetMostChallengingAssessments(limit int) ([]map[string]interface{}, error) {
	args := m.Called(limit)
	results, _ := args.Get(0).([]map[string]interface{})
	return results, args.Error(1)
}

func (m *MockAttemptRepository) GetMostSuccessfulAssessments(limit int) ([]map[string]interface{}, error) {
	args := m.Called(limit)
	results, _ := args.Get(0).([]map[string]interface{})
	return results, args.Error(1)
}

func (m *MockAttemptRepository) GetPassRate() (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockAttemptRepository) CountAll() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAttemptRepository) CountByPeriod(days int) (int64, error) {
	args := m.Called(days)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAttemptRepository) SaveSuspiciousActivity(activity *models.SuspiciousActivity) error {
	args := m.Called(activity)
	return args.Error(0)
}

func (m *MockAttemptRepository) CountRecentSuspiciousActivity(hours int) (int64, error) {
	args := m.Called(hours)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAttemptRepository) FindSuspiciousActivitiesByAttemptID(attemptID uint) ([]models.SuspiciousActivity, error) {
	args := m.Called(attemptID)
	activities, _ := args.Get(0).([]models.SuspiciousActivity)
	return activities, args.Error(1)
}

func (m *MockAttemptRepository) ExpiredAttempt() ([]models.Attempt, error) {
	args := m.Called()
	attempts, _ := args.Get(0).([]models.Attempt)
	return attempts, args.Error(1)
}

func (m *MockAttemptRepository) IsUserInAttempt(userID uint) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

// --- Test Cases ---

func TestAttemptService_GetListAttemptByUserAndAssessment(t *testing.T) {
	mockRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAttemptService(mockRepo, logger)

	userID := uint(1)
	assessmentID := uint(10)
	params := util.PaginationParams{Page: 0, Limit: 10}
	expectedAttempts := []models.Attempt{{ID: 1, UserID: userID, AssessmentID: assessmentID}}
	expectedTotal := int64(1)

	mockRepo.On("ListAttemptByUserAndAssessmentID", userID, assessmentID, params).Return(expectedAttempts, expectedTotal, nil)

	attempts, total, err := service.GetListAttemptByUserAndAssessment(userID, assessmentID, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedAttempts, attempts)
	assert.Equal(t, expectedTotal, total)
	mockRepo.AssertExpectations(t)
}

func TestAttemptService_GetListAttemptByUserAndAssessment_RepoError(t *testing.T) {
	mockRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAttemptService(mockRepo, logger)

	userID := uint(1)
	assessmentID := uint(10)
	params := util.PaginationParams{Page: 0, Limit: 10}
	repoError := errors.New("database error")

	mockRepo.On("ListAttemptByUserAndAssessmentID", userID, assessmentID, params).Return(nil, int64(0), repoError)

	attempts, total, err := service.GetListAttemptByUserAndAssessment(userID, assessmentID, params)

	assert.Error(t, err)
	assert.Nil(t, attempts)
	assert.Zero(t, total)
	assert.Equal(t, repoError, err)
	mockRepo.AssertExpectations(t)
}

func TestAttemptService_GetAttemptDetail(t *testing.T) {
	mockRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAttemptService(mockRepo, logger)

	attemptID := uint(5)
	expectedAttempt := &models.Attempt{ID: attemptID, Status: "Completed"}

	mockRepo.On("FindByID", attemptID).Return(expectedAttempt, nil)

	attempt, err := service.GetAttemptDetail(attemptID)

	assert.NoError(t, err)
	assert.Equal(t, expectedAttempt, attempt)
	mockRepo.AssertExpectations(t)
}

func TestAttemptService_GetAttemptDetail_NotFound(t *testing.T) {
	mockRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAttemptService(mockRepo, logger)

	attemptID := uint(99)
	repoError := errors.New("record not found")

	mockRepo.On("FindByID", attemptID).Return(nil, repoError)

	attempt, err := service.GetAttemptDetail(attemptID)

	assert.Error(t, err)
	assert.Nil(t, attempt)
	assert.Equal(t, repoError, err)
	mockRepo.AssertExpectations(t)
}

func TestAttemptService_GradeAttempt(t *testing.T) {
	mockRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAttemptService(mockRepo, logger)

	attemptID := uint(1)
	newScore := 85.5
	newFeedback := "Good job!"

	gradeData := models.AttemptUpdateDTO{
		Score:    newScore,
		Feedback: newFeedback,
		Answers: []struct {
			ID        uint `json:"id"`
			IsCorrect bool `json:"isCorrect"`
		}{
			{ID: 10, IsCorrect: true},
			{ID: 11, IsCorrect: false},
		},
	}

	existingAttempt := &models.Attempt{
		ID:       attemptID,
		Score:    nil, // Chưa có điểm
		Feedback: "",  // Chưa có feedback
		Answers: []models.Answer{
			{ID: 10, AttemptID: attemptID, QuestionID: 101, IsCorrect: nil}, // Chưa chấm
			{ID: 11, AttemptID: attemptID, QuestionID: 102, IsCorrect: nil}, // Chưa chấm
			{ID: 12, AttemptID: attemptID, QuestionID: 103, IsCorrect: nil}, // Câu này không được chấm lại
		},
	}

	// Expect FindByID to be called
	mockRepo.On("FindByID", attemptID).Return(existingAttempt, nil)

	// Expect Update to be called with the modified attempt
	mockRepo.On("Update", mock.MatchedBy(func(a *models.Attempt) bool {
		// Kiểm tra điểm và feedback
		scoreMatch := a.Score != nil && *a.Score == newScore
		feedbackMatch := a.Feedback == newFeedback
		// Kiểm tra các câu trả lời đã được cập nhật isCorrect
		answer10Correct := false
		answer11Correct := false
		answer12Unchanged := false
		for _, ans := range a.Answers {
			if ans.ID == 10 && ans.IsCorrect != nil && *ans.IsCorrect == true {
				answer10Correct = true
			}
			if ans.ID == 11 && ans.IsCorrect != nil && *ans.IsCorrect == false {
				answer11Correct = true
			}
			if ans.ID == 12 && ans.IsCorrect == nil { // Đảm bảo câu 12 không bị thay đổi
				answer12Unchanged = true
			}
		}
		return scoreMatch && feedbackMatch && answer10Correct && answer11Correct && answer12Unchanged
	})).Return(nil)

	err := service.GradeAttempt(gradeData, attemptID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAttemptService_GradeAttempt_AttemptNotFound(t *testing.T) {
	mockRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAttemptService(mockRepo, logger)

	attemptID := uint(99)
	gradeData := models.AttemptUpdateDTO{} // Dữ liệu không quan trọng vì sẽ lỗi trước đó

	mockRepo.On("FindByID", attemptID).Return(nil, errors.New("record not found"))

	err := service.GradeAttempt(gradeData, attemptID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything) // Đảm bảo Update không được gọi
}

func TestAttemptService_GradeAttempt_UpdateError(t *testing.T) {
	mockRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAttemptService(mockRepo, logger)

	attemptID := uint(1)
	gradeData := models.AttemptUpdateDTO{Score: 90.0}
	existingAttempt := &models.Attempt{ID: attemptID, Answers: []models.Answer{}} // Attempt rỗng để đơn giản

	mockRepo.On("FindByID", attemptID).Return(existingAttempt, nil)
	mockRepo.On("Update", mock.Anything).Return(errors.New("db update error")) // Giả lập lỗi khi Update

	err := service.GradeAttempt(gradeData, attemptID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db update error")
	mockRepo.AssertExpectations(t) // Cả FindByID và Update đều được gọi
}
