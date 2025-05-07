package service

import (
	// Không import mock repo nữa
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// --- Mock AssessmentRepository ---
type MockAssessmentRepository struct{ mock.Mock }

func (m *MockAssessmentRepository) Create(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}
func (m *MockAssessmentRepository) FindByID(id uint) (*models.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Assessment), args.Error(1)
}
func (m *MockAssessmentRepository) Update(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}
func (m *MockAssessmentRepository) Delete(id uint) error { args := m.Called(id); return args.Error(0) }
func (m *MockAssessmentRepository) List(params util.PaginationParams) ([]models.Assessment, int64, error) {
	args := m.Called(params)
	return args.Get(0).([]models.Assessment), args.Get(1).(int64), args.Error(2)
}
func (m *MockAssessmentRepository) FindRecent(limit int) ([]models.Assessment, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Assessment), args.Error(1)
}
func (m *MockAssessmentRepository) GetStatistics() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAssessmentRepository) UpdateSettings(id uint, settings *models.AssessmentSettings) error {
	args := m.Called(id, settings)
	return args.Error(0)
}
func (m *MockAssessmentRepository) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(id, params)
	return args.Get(0).([]map[string]interface{}), args.Get(1).(int64), args.Error(2)
}
func (m *MockAssessmentRepository) Publish(id uint) error { args := m.Called(id); return args.Error(0) }
func (m *MockAssessmentRepository) Duplicate(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}
func (m *MockAssessmentRepository) GetAssessmentHasAttemptByUser(params util.PaginationParams, userID uint) ([]models.Assessment, int64, error) {
	args := m.Called(params, userID)
	return args.Get(0).([]models.Assessment), args.Get(1).(int64), args.Error(2)
}

// --- Mock AttemptRepository ---
type MockAttemptRepository struct{ mock.Mock }

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

// Thêm các hàm mock còn thiếu nếu cần

// --- Mock QuestionRepository ---
type MockQuestionRepository struct{ mock.Mock }

func (m *MockQuestionRepository) Create(question *models.Question) error {
	args := m.Called(question)
	return args.Error(0)
}
func (m *MockQuestionRepository) FindByID(id uint) (*models.Question, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Question), args.Error(1)
}
func (m *MockQuestionRepository) FindByAssessmentID(assessmentID uint) ([]models.Question, error) {
	args := m.Called(assessmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Question), args.Error(1)
}
func (m *MockQuestionRepository) Update(question *models.Question) error {
	args := m.Called(question)
	return args.Error(0)
}
func (m *MockQuestionRepository) Delete(id uint) error { args := m.Called(id); return args.Error(0) }
func (m *MockQuestionRepository) AddOption(option *models.QuestionOption) error {
	args := m.Called(option)
	return args.Error(0)
}
func (m *MockQuestionRepository) UpdateOption(option *models.QuestionOption) error {
	args := m.Called(option)
	return args.Error(0)
}
func (m *MockQuestionRepository) DeleteOption(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// --- Mock UserRepository ---
type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockUserRepository) Delete(id uint) error { args := m.Called(id); return args.Error(0) }
func (m *MockUserRepository) List(params util.PaginationParams) ([]models.User, int64, error) {
	args := m.Called(params)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}
func (m *MockUserRepository) UpdateLastLogin(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockUserRepository) GetUserStats() (int64, int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}
func (m *MockUserRepository) CountAll() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockUserRepository) GetNewUsersCount(days int) (int64, error) {
	args := m.Called(days)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockUserRepository) GetListUserByAssessment(params util.PaginationParams, assessmentID uint) ([]models.User, int64, error) {
	args := m.Called(params, assessmentID)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

// --- Test Cases ---

func TestStudentService_GetAvailableAssessments(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	mockUserRepo := new(MockUserRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(mockAssessmentRepo, mockAttemptRepo, mockQuestionRepo, mockUserRepo, logger)

	userID := uint(1)
	params := util.PaginationParams{Page: 0, Limit: 10}
	expectedAssessments := []map[string]interface{}{{"id": "1", "title": "Available Quiz"}}
	expectedTotal := int64(1)

	mockUserRepo.On("FindByID", userID).Return(&models.User{ID: userID}, nil)
	mockAttemptRepo.On("FindAvailableAssessments", userID, params).Return(expectedAssessments, expectedTotal, nil)

	assessments, total, err := service.GetAvailableAssessments(userID, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedAssessments, assessments)
	assert.Equal(t, expectedTotal, total)
	mockUserRepo.AssertExpectations(t)
	mockAttemptRepo.AssertExpectations(t)
}

func TestStudentService_GetAvailableAssessments_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(nil, mockAttemptRepo, nil, mockUserRepo, logger)

	userID := uint(99)
	params := util.PaginationParams{}

	mockUserRepo.On("FindByID", userID).Return(nil, errors.New("user not found"))

	assessments, total, err := service.GetAvailableAssessments(userID, params)

	assert.Error(t, err)
	assert.Nil(t, assessments)
	assert.Zero(t, total)
	assert.Contains(t, err.Error(), "user not found")
	mockUserRepo.AssertExpectations(t)
	mockAttemptRepo.AssertNotCalled(t, "FindAvailableAssessments", mock.Anything, mock.Anything)
}

func TestStudentService_StartAssessment_Success(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	mockUserRepo := new(MockUserRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(mockAssessmentRepo, mockAttemptRepo, mockQuestionRepo, mockUserRepo, logger)

	userID := uint(1)
	assessmentID := uint(10)
	assessment := &models.Assessment{
		ID:       assessmentID,
		Status:   "active",
		Duration: 60,
		Settings: models.AssessmentSettings{
			AllowRetake:        false, // Chỉ được làm 1 lần
			MaxAttempts:        1,
			RandomizeQuestions: false, // Không random
		},
	}
	questions := []models.Question{
		{ID: 101, AssessmentID: assessmentID, Text: "Q1", CorrectAnswer: "A"},
		{ID: 102, AssessmentID: assessmentID, Text: "Q2", CorrectAnswer: "B"},
	}
	expectedAttempt := &models.Attempt{UserID: userID, AssessmentID: assessmentID, Status: "In Progress"} // Trạng thái sau khi tạo

	mockUserRepo.On("FindByID", userID).Return(&models.User{ID: userID}, nil)
	mockAssessmentRepo.On("FindByID", assessmentID).Return(assessment, nil)
	mockAttemptRepo.On("IsUserInAttempt", userID).Return(false, nil)                      // User không đang làm bài
	mockAttemptRepo.On("HasCompletedAssessment", userID, assessmentID).Return(false, nil) // User chưa hoàn thành bài này
	// Expect Create attempt được gọi
	mockAttemptRepo.On("Create", mock.MatchedBy(func(att *models.Attempt) bool {
		return att.UserID == userID && att.AssessmentID == assessmentID && att.Status == "In Progress"
	})).Return(nil).Run(func(args mock.Arguments) {
		// Gán ID giả lập sau khi tạo
		att := args.Get(0).(*models.Attempt)
		att.ID = 999
		expectedAttempt.ID = 999                  // Cập nhật ID cho expected
		expectedAttempt.StartedAt = att.StartedAt // Cập nhật thời gian bắt đầu
	})
	mockQuestionRepo.On("FindByAssessmentID", assessmentID).Return(questions, nil)

	attempt, studentQuestions, settings, returnedAssessment, err := service.StartAssessment(userID, assessmentID)

	assert.NoError(t, err)
	require.NotNil(t, attempt)
	assert.Equal(t, expectedAttempt.ID, attempt.ID)
	assert.Equal(t, "In Progress", attempt.Status)
	assert.Equal(t, assessment.Settings, *settings)
	assert.Equal(t, assessment, returnedAssessment)
	require.Len(t, studentQuestions, 2)
	assert.Empty(t, studentQuestions[0].CorrectAnswer) // Đảm bảo đáp án đúng đã bị xóa
	assert.Empty(t, studentQuestions[1].CorrectAnswer)

	mockUserRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertExpectations(t)
	mockAttemptRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

// Thêm các test case lỗi cho StartAssessment:
// - Assessment không tìm thấy
// - Assessment không active
// - Assessment quá hạn
// - User đang làm bài khác (IsUserInAttempt = true)
// - Hết lượt làm bài (AllowRetake = true, CountAttempts >= MaxAttempts)
// - Đã hoàn thành bài (AllowRetake = false, HasCompleted = true)
// - Lỗi khi tạo Attempt
// - Lỗi khi lấy Questions

func TestStudentService_GetAssessmentResultsHistory(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	mockUserRepo := new(MockUserRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(mockAssessmentRepo, mockAttemptRepo, nil, mockUserRepo, logger)

	userID := uint(1)
	assessmentID := uint(10)
	expectedResults := []map[string]interface{}{{"attemptId": 1, "score": 80.0}}

	mockUserRepo.On("FindByID", userID).Return(&models.User{ID: userID}, nil)
	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	mockAttemptRepo.On("FindCompletedAttemptsByUserAndAssessment", userID, assessmentID).Return(expectedResults, nil)

	results, err := service.GetAssessmentResultsHistory(userID, assessmentID)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
	mockUserRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertExpectations(t)
	mockAttemptRepo.AssertExpectations(t)
}

// Thêm test case lỗi cho GetAssessmentResultsHistory

func TestStudentService_GetAttemptDetails(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(mockAssessmentRepo, mockAttemptRepo, nil, nil, logger) // Không cần UserRepo

	attemptID := uint(5)
	userID := uint(1)
	assessmentID := uint(10)
	startTime := time.Now().Add(-30 * time.Minute)
	attempt := &models.Attempt{ID: attemptID, UserID: userID, AssessmentID: assessmentID, StartedAt: startTime, Status: "In Progress", Answers: []models.Answer{{QuestionID: 101}}}
	assessment := &models.Assessment{ID: assessmentID, Duration: 60, Questions: []models.Question{{ID: 101}, {ID: 102}}} // 2 câu hỏi

	mockAttemptRepo.On("FindByID", attemptID).Return(attempt, nil)
	mockAssessmentRepo.On("FindByID", assessmentID).Return(assessment, nil)

	details, err := service.GetAttemptDetails(attemptID, userID)

	assert.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, (attemptID), (*details)["attemptId"]) // JSON number
	assert.Equal(t, "In Progress", (*details)["status"])
	progress, ok := (*details)["progress"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, int(1), progress["answered"])    // 1 câu đã trả lời
	assert.Equal(t, int(2), progress["total"])       // Tổng 2 câu
	assert.Equal(t, int(50), progress["percentage"]) // 50%

	mockAttemptRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertExpectations(t)
}

// Thêm test case lỗi cho GetAttemptDetails (attempt not found, unauthorized, assessment not found)

func TestStudentService_SaveAnswer_New(t *testing.T) {
	mockAttemptRepo := new(MockAttemptRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(nil, mockAttemptRepo, mockQuestionRepo, nil, logger)

	attemptID := uint(1)
	questionID := uint(101)
	userID := uint(5)
	answerStr := "true"
	attempt := &models.Attempt{ID: attemptID, UserID: userID, AssessmentID: 10, Status: "In Progress"}
	question := &models.Question{ID: questionID, AssessmentID: 10, Type: "true-false", CorrectAnswer: "true", Points: 1}
	expectedAnswer := &models.Answer{AttemptID: attemptID, QuestionID: questionID, Answer: answerStr, IsCorrect: &[]bool{true}[0]} // Correct answer

	mockAttemptRepo.On("FindByID", attemptID).Return(attempt, nil)
	mockQuestionRepo.On("FindByID", questionID).Return(question, nil)
	// Expect FindAnswerByAttemptAndQuestion to return not found (nil, nil)
	mockAttemptRepo.On("FindAnswerByAttemptAndQuestion", attemptID, questionID).Return(nil, nil) // Simulate answer not existing
	// Expect SaveAnswer to be called
	mockAttemptRepo.On("SaveAnswer", mock.MatchedBy(func(ans *models.Answer) bool {
		return ans.AttemptID == expectedAnswer.AttemptID &&
			ans.QuestionID == expectedAnswer.QuestionID &&
			ans.Answer == expectedAnswer.Answer &&
			ans.IsCorrect != nil && *ans.IsCorrect == *expectedAnswer.IsCorrect
	})).Return(nil)

	err := service.SaveAnswer(attemptID, questionID, answerStr, userID)

	assert.NoError(t, err)
	mockAttemptRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

func TestStudentService_SaveAnswer_Update(t *testing.T) {
	mockAttemptRepo := new(MockAttemptRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(nil, mockAttemptRepo, mockQuestionRepo, nil, logger)

	attemptID := uint(1)
	questionID := uint(101)
	userID := uint(5)
	newAnswerStr := "false"
	attempt := &models.Attempt{ID: attemptID, UserID: userID, AssessmentID: 10, Status: "In Progress"}
	question := &models.Question{ID: questionID, AssessmentID: 10, Type: "true-false", CorrectAnswer: "true", Points: 1}
	existingAnswer := &models.Answer{ID: 50, AttemptID: attemptID, QuestionID: questionID, Answer: "true", IsCorrect: &[]bool{true}[0]}
	expectedUpdatedAnswer := &models.Answer{ID: 50, AttemptID: attemptID, QuestionID: questionID, Answer: newAnswerStr, IsCorrect: &[]bool{false}[0]} // Incorrect answer now

	mockAttemptRepo.On("FindByID", attemptID).Return(attempt, nil)
	mockQuestionRepo.On("FindByID", questionID).Return(question, nil)
	// Expect FindAnswerByAttemptAndQuestion to return the existing answer
	mockAttemptRepo.On("FindAnswerByAttemptAndQuestion", attemptID, questionID).Return(existingAnswer, nil)
	// Expect UpdateAnswer to be called
	mockAttemptRepo.On("UpdateAnswer", mock.MatchedBy(func(ans *models.Answer) bool {
		return ans.ID == expectedUpdatedAnswer.ID &&
			ans.Answer == expectedUpdatedAnswer.Answer &&
			ans.IsCorrect != nil && *ans.IsCorrect == *expectedUpdatedAnswer.IsCorrect
	})).Return(nil)

	err := service.SaveAnswer(attemptID, questionID, newAnswerStr, userID)

	assert.NoError(t, err)
	mockAttemptRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
	mockAttemptRepo.AssertNotCalled(t, "SaveAnswer", mock.Anything) // SaveAnswer should not be called
}

// Thêm test case lỗi cho SaveAnswer (attempt not found, unauthorized, attempt not in progress, question not found, question not in assessment, invalid answer format)

func TestStudentService_SubmitAssessment(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(mockAssessmentRepo, mockAttemptRepo, mockQuestionRepo, nil, logger)

	attemptID := uint(1)
	userID := uint(5)
	assessmentID := uint(10)
	startTime := time.Now().Add(-20 * time.Minute) // Started 20 mins ago
	correct := true
	// incorrect := false

	attempt := &models.Attempt{
		ID:           attemptID,
		UserID:       userID,
		AssessmentID: assessmentID,
		StartedAt:    startTime,
		Status:       "In Progress",
		Answers: []models.Answer{
			{QuestionID: 101, Answer: "true", IsCorrect: &correct},  // Correct
			{QuestionID: 102, Answer: "Essay text", IsCorrect: nil}, // Essay, needs grading
			// Question 103 is unanswered
		},
	}
	assessment := &models.Assessment{
		ID:           assessmentID,
		PassingScore: 50.0,
		Duration:     30, // 30 min duration
		Settings:     models.AssessmentSettings{ShowResults: true},
	}
	questions := []models.Question{
		{ID: 101, AssessmentID: assessmentID, Points: 10, Type: "true-false", CorrectAnswer: "true"},
		{ID: 102, AssessmentID: assessmentID, Points: 15, Type: "essay"},
		{ID: 103, AssessmentID: assessmentID, Points: 5, Type: "multiple-choice", CorrectAnswer: "a"}, // Unanswered
	}
	totalPoints := 10.0 + 15.0 + 5.0                    // 30
	earnedPoints := 10.0                                // Only from Q101
	expectedScore := (earnedPoints / totalPoints) * 100 // ~33.33
	expectedStatus := "Failed"                          // Below passing score

	mockAttemptRepo.On("FindByID", attemptID).Return(attempt, nil)
	mockAssessmentRepo.On("FindByID", assessmentID).Return(assessment, nil)
	mockQuestionRepo.On("FindByAssessmentID", assessmentID).Return(questions, nil)
	// Expect Update attempt được gọi với điểm và status đã tính
	mockAttemptRepo.On("Update", mock.MatchedBy(func(att *models.Attempt) bool {
		scoreMatch := att.Score != nil && *att.Score >= expectedScore-0.01 && *att.Score <= expectedScore+0.01 // Check score with tolerance
		statusMatch := att.Status == expectedStatus
		submittedMatch := att.SubmittedAt != nil
		endedMatch := att.EndedAt != nil
		durationMatch := att.Duration != nil && *att.Duration >= 20 // Duration should be around 20 mins
		return scoreMatch && statusMatch && submittedMatch && endedMatch && durationMatch
	})).Return(nil)

	result, err := service.SubmitAssessment(attemptID, userID)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, (attemptID), (*result)["attemptId"])
	assert.True(t, (*result)["completed"].(bool))
	resultsMap, ok := (*result)["results"].(map[string]interface{})
	require.True(t, ok)
	assert.InDelta(t, expectedScore, resultsMap["score"], 0.01)
	assert.Equal(t, (3), resultsMap["totalQuestions"])
	assert.Equal(t, (1), resultsMap["correctAnswers"])
	assert.Equal(t, (0), resultsMap["incorrectAnswers"]) // Q102 is essay, Q103 unanswered
	assert.Equal(t, (1), resultsMap["unanswered"])
	assert.Equal(t, (1), resultsMap["essayQuestions"])
	assert.Equal(t, expectedStatus, resultsMap["status"])

	mockAttemptRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

// Thêm test case lỗi cho SubmitAssessment

func TestStudentService_SubmitMonitorEvent(t *testing.T) {
	mockAttemptRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(nil, mockAttemptRepo, nil, nil, logger)

	attemptID := uint(1)
	userID := uint(5)
	attempt := &models.Attempt{ID: attemptID, UserID: userID, AssessmentID: 10, Status: "In Progress"}
	eventType := "TAB_SWITCH"
	details := map[string]interface{}{"count": 3.0}

	mockAttemptRepo.On("FindByID", attemptID).Return(attempt, nil)
	// Expect SaveSuspiciousActivity to be called
	mockAttemptRepo.On("SaveSuspiciousActivity", mock.MatchedBy(func(sa *models.SuspiciousActivity) bool {
		return sa.AttemptID == attemptID &&
			sa.UserID == userID &&
			sa.Type == eventType &&
			sa.Severity == "CRITICAL" // Severity for TAB_SWITCH
	})).Return(nil)

	result, err := service.SubmitMonitorEvent(attemptID, eventType, details, nil, userID) // No image data

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, (*result)["received"].(bool))
	assert.Equal(t, "CRITICAL", (*result)["severity"])
	assert.NotEmpty(t, (*result)["message"])

	mockAttemptRepo.AssertExpectations(t)
}

// Thêm test case lỗi cho SubmitMonitorEvent

func TestStudentService_GetAllAttemptByUserID(t *testing.T) {
	mockAttemptRepo := new(MockAttemptRepository)
	mockUserRepo := new(MockUserRepository)
	logger := zaptest.NewLogger(t)
	service := NewStudentService(nil, mockAttemptRepo, nil, mockUserRepo, logger)

	userID := uint(1)
	params := util.PaginationParams{Limit: 5}
	expectedAttempts := []models.Attempt{{ID: 1}, {ID: 2}}
	expectedTotal := int64(2)

	mockUserRepo.On("FindByID", userID).Return(&models.User{ID: userID}, nil)
	mockAttemptRepo.On("GetAllAttemptByUserId", userID, params).Return(expectedAttempts, expectedTotal, nil)

	attempts, total, err := service.GetAllAttemptByUserID(userID, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedAttempts, attempts)
	assert.Equal(t, expectedTotal, total)
	mockUserRepo.AssertExpectations(t)
	mockAttemptRepo.AssertExpectations(t)
}

// Thêm test case lỗi cho GetAllAttemptByUserID (user not found, repo error)

// Test cho AutoSubmitAssessment phức tạp hơn vì nó liên quan đến thời gian và nhiều bước,
// có thể phù hợp hơn với integration test hoặc test thủ công.
// Tuy nhiên, có thể viết unit test bằng cách mock ExpiredAttempt, FindByID, FindByAssessmentID, Update.
