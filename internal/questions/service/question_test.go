package service

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock QuestionRepository ---
type MockQuestionRepository struct {
	mock.Mock
}

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

// --- Mock AssessmentRepository ---
type MockAssessmentRepository struct {
	mock.Mock
}

// Implement AssessmentRepository interface for mock
func (m *MockAssessmentRepository) Create(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}

func (m *MockAssessmentRepository) FindByID(id uint) (*models.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Ensure the returned object is of the correct type
	assessment, ok := args.Get(0).(*models.Assessment)
	if !ok && args.Get(0) != nil {
		// Handle cases where the mock might return something unexpected but not nil
		panic("Mock FindByID returned non-nil value of incorrect type")
	}
	return assessment, args.Error(1)
}

func (m *MockAssessmentRepository) Update(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}

func (m *MockAssessmentRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAssessmentRepository) List(params util.PaginationParams) ([]models.Assessment, int64, error) {
	args := m.Called(params)
	// Ensure correct type assertion for slice
	assessments, ok := args.Get(0).([]models.Assessment)
	if !ok && args.Get(0) != nil {
		panic("Mock List returned non-nil value of incorrect type for assessments")
	}
	// Ensure correct type assertion for int64
	count, ok := args.Get(1).(int64)
	if !ok {
		// Try converting from int if necessary, though int64 is expected
		if intCount, okInt := args.Get(1).(int); okInt {
			count = int64(intCount)
		} else {
			panic("Mock List returned value of incorrect type for count")
		}
	}
	return assessments, count, args.Error(2)
}

func (m *MockAssessmentRepository) FindRecent(limit int) ([]models.Assessment, error) {
	args := m.Called(limit)
	assessments, ok := args.Get(0).([]models.Assessment)
	if !ok && args.Get(0) != nil {
		panic("Mock FindRecent returned non-nil value of incorrect type")
	}
	return assessments, args.Error(1)
}

func (m *MockAssessmentRepository) GetStatistics() (map[string]interface{}, error) {
	args := m.Called()
	stats, ok := args.Get(0).(map[string]interface{})
	if !ok && args.Get(0) != nil {
		panic("Mock GetStatistics returned non-nil value of incorrect type")
	}
	return stats, args.Error(1)
}

func (m *MockAssessmentRepository) UpdateSettings(id uint, settings *models.AssessmentSettings) error {
	args := m.Called(id, settings)
	return args.Error(0)
}

func (m *MockAssessmentRepository) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(id, params)
	results, ok := args.Get(0).([]map[string]interface{})
	if !ok && args.Get(0) != nil {
		panic("Mock GetResults returned non-nil value of incorrect type for results")
	}
	count, ok := args.Get(1).(int64)
	if !ok {
		if intCount, okInt := args.Get(1).(int); okInt {
			count = int64(intCount)
		} else {
			panic("Mock GetResults returned value of incorrect type for count")
		}
	}
	return results, count, args.Error(2)
}

func (m *MockAssessmentRepository) Publish(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAssessmentRepository) Duplicate(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}

func (m *MockAssessmentRepository) GetAssessmentHasAttemptByUser(params util.PaginationParams, userID uint) ([]models.Assessment, int64, error) {
	args := m.Called(params, userID)
	assessments, ok := args.Get(0).([]models.Assessment)
	if !ok && args.Get(0) != nil {
		panic("Mock GetAssessmentHasAttemptByUser returned non-nil value of incorrect type for assessments")
	}
	count, ok := args.Get(1).(int64)
	if !ok {
		if intCount, okInt := args.Get(1).(int); okInt {
			count = int64(intCount)
		} else {
			panic("Mock GetAssessmentHasAttemptByUser returned value of incorrect type for count")
		}
	}
	return assessments, count, args.Error(2)
}

// Mock UserRepository implementation (assuming it exists based on service dependencies)
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	user, ok := args.Get(0).(*models.User)
	if !ok && args.Get(0) != nil {
		panic("Mock FindByID returned non-nil value of incorrect type")
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	user, ok := args.Get(0).(*models.User)
	if !ok && args.Get(0) != nil {
		panic("Mock FindByEmail returned non-nil value of incorrect type")
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(params util.PaginationParams) ([]models.User, int64, error) {
	args := m.Called(params)
	users, ok := args.Get(0).([]models.User)
	if !ok && args.Get(0) != nil {
		panic("Mock List returned non-nil value of incorrect type for users")
	}
	count, ok := args.Get(1).(int64)
	if !ok {
		if intCount, okInt := args.Get(1).(int); okInt {
			count = int64(intCount)
		} else {
			panic("Mock List returned value of incorrect type for count")
		}
	}
	return users, count, args.Error(2)
}

func (m *MockUserRepository) UpdateLastLogin(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserStats() (int64, int64, error) {
	args := m.Called()
	active, ok1 := args.Get(0).(int64)
	inactive, ok2 := args.Get(1).(int64)
	if !ok1 || !ok2 {
		panic("Mock GetUserStats returned value of incorrect type")
	}
	return active, inactive, args.Error(2)
}

func (m *MockUserRepository) CountAll() (int64, error) {
	args := m.Called()
	count, ok := args.Get(0).(int64)
	if !ok {
		panic("Mock CountAll returned value of incorrect type")
	}
	return count, args.Error(1)
}

func (m *MockUserRepository) GetNewUsersCount(days int) (int64, error) {
	args := m.Called(days)
	count, ok := args.Get(0).(int64)
	if !ok {
		panic("Mock GetNewUsersCount returned value of incorrect type")
	}
	return count, args.Error(1)
}

func (m *MockUserRepository) GetListUserByAssessment(params util.PaginationParams, assessmentID uint) ([]models.User, int64, error) {
	args := m.Called(params, assessmentID)
	users, ok := args.Get(0).([]models.User)
	if !ok && args.Get(0) != nil {
		panic("Mock GetListUserByAssessment returned non-nil value of incorrect type for users")
	}
	count, ok := args.Get(1).(int64)
	if !ok {
		if intCount, okInt := args.Get(1).(int); okInt {
			count = int64(intCount)
		} else {
			panic("Mock GetListUserByAssessment returned value of incorrect type for count")
		}
	}
	return users, count, args.Error(2)
}

// --- Test Cases ---

func TestQuestionService_AddQuestion_Success_MC(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{
		Type:          "multiple-choice",
		Text:          "MC Question?",
		CorrectAnswer: "a",
		Points:        5,
		Options: []models.QuestionOption{
			{OptionID: "a", Text: "Option A"},
			{OptionID: "b", Text: "Option B"},
		},
	}
	createdQuestion := &models.Question{ID: 10, AssessmentID: assessmentID, Type: "multiple-choice", Text: "MC Question?", CorrectAnswer: "a", Points: 5, Options: question.Options}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	mockQuestionRepo.On("Create", question).Return(nil).Run(func(args mock.Arguments) {
		q := args.Get(0).(*models.Question)
		q.ID = 10 // Simulate ID assignment by repo
	})
	// Giả lập repo trả về question đã tạo với ID
	mockQuestionRepo.On("FindByID", uint(10)).Return(createdQuestion, nil)

	result, err := service.AddQuestion(assessmentID, question)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(10), result.ID)
	assert.Equal(t, assessmentID, result.AssessmentID)
	assert.Equal(t, "multiple-choice", result.Type)
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

func TestQuestionService_AddQuestion_Success_TF(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{
		Type:          "true-false",
		Text:          "TF Question?",
		CorrectAnswer: "true",
		Points:        2,
	}
	createdQuestion := &models.Question{ID: 11, AssessmentID: assessmentID, Type: "true-false", Text: "TF Question?", CorrectAnswer: "true", Points: 2}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	mockQuestionRepo.On("Create", question).Return(nil).Run(func(args mock.Arguments) { args.Get(0).(*models.Question).ID = 11 })
	mockQuestionRepo.On("FindByID", uint(11)).Return(createdQuestion, nil)

	result, err := service.AddQuestion(assessmentID, question)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(11), result.ID)
	assert.Equal(t, "true-false", result.Type)
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

func TestQuestionService_AddQuestion_Success_Essay(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{
		Type:   "essay",
		Text:   "Essay Question?",
		Points: 15,
	}
	// Service sẽ tự set CorrectAnswer thành "" cho essay
	expectedQuestionToCreate := &models.Question{
		AssessmentID:  assessmentID,
		Type:          "essay",
		Text:          "Essay Question?",
		CorrectAnswer: "", // Service sẽ set cái này
		Points:        15,
	}
	createdQuestion := &models.Question{ID: 12, AssessmentID: assessmentID, Type: "essay", Text: "Essay Question?", CorrectAnswer: "", Points: 15}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	// Kiểm tra xem CorrectAnswer có được set thành "" không
	mockQuestionRepo.On("Create", mock.MatchedBy(func(q *models.Question) bool {
		return q.Type == expectedQuestionToCreate.Type &&
			q.Text == expectedQuestionToCreate.Text &&
			q.CorrectAnswer == expectedQuestionToCreate.CorrectAnswer &&
			q.Points == expectedQuestionToCreate.Points &&
			q.AssessmentID == expectedQuestionToCreate.AssessmentID
	})).Return(nil).Run(func(args mock.Arguments) { args.Get(0).(*models.Question).ID = 12 })
	mockQuestionRepo.On("FindByID", uint(12)).Return(createdQuestion, nil)

	result, err := service.AddQuestion(assessmentID, question)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(12), result.ID)
	assert.Equal(t, "essay", result.Type)
	assert.Empty(t, result.CorrectAnswer) // Đảm bảo CorrectAnswer là rỗng
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

func TestQuestionService_AddQuestion_AssessmentNotFound(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(99)
	question := &models.Question{Type: "essay", Text: "Test"}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(nil, errors.New("assessment not found"))

	result, err := service.AddQuestion(assessmentID, question)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assessment not found")
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Create", mock.Anything) // Create không được gọi
}

func TestQuestionService_AddQuestion_InvalidType(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{Type: "invalid-type", Text: "Test"}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)

	result, err := service.AddQuestion(assessmentID, question)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid question type")
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestQuestionService_AddQuestion_MC_NoOptions(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{Type: "multiple-choice", Text: "Test", Options: []models.QuestionOption{}} // No options

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)

	result, err := service.AddQuestion(assessmentID, question)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "multiple-choice questions require options")
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestQuestionService_AddQuestion_MC_InvalidCorrectAnswer(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{
		Type:          "multiple-choice",
		Text:          "Test",
		CorrectAnswer: "c", // 'c' không có trong options
		Options: []models.QuestionOption{
			{OptionID: "a", Text: "A"},
			{OptionID: "b", Text: "B"},
		},
	}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)

	result, err := service.AddQuestion(assessmentID, question)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "correct answer must match one of the option IDs")
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestQuestionService_AddQuestion_TF_InvalidCorrectAnswer(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{Type: "true-false", Text: "Test", CorrectAnswer: "maybe"} // Invalid

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)

	result, err := service.AddQuestion(assessmentID, question)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "correct answer for true-false questions must be 'true' or 'false'")
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestQuestionService_AddQuestion_RepoCreateError(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	question := &models.Question{Type: "essay", Text: "Test"}
	repoError := errors.New("db create error")

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	mockQuestionRepo.On("Create", question).Return(repoError)

	result, err := service.AddQuestion(assessmentID, question)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repoError, err)
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
	// FindByID không được gọi nếu Create lỗi
	mockQuestionRepo.AssertNotCalled(t, "FindByID", mock.Anything)
}

func TestQuestionService_GetQuestionsByAssessment(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	expectedQuestions := []models.Question{
		{ID: 1, AssessmentID: assessmentID, Text: "Q1"},
		{ID: 2, AssessmentID: assessmentID, Text: "Q2"},
	}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	mockQuestionRepo.On("FindByAssessmentID", assessmentID).Return(expectedQuestions, nil)

	result, err := service.GetQuestionsByAssessment(assessmentID)

	assert.NoError(t, err)
	assert.Equal(t, expectedQuestions, result)
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

func TestQuestionService_GetQuestionsByAssessment_AssessmentNotFound(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(99)

	mockAssessmentRepo.On("FindByID", assessmentID).Return(nil, errors.New("assessment not found"))

	result, err := service.GetQuestionsByAssessment(assessmentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assessment not found")
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "FindByAssessmentID", mock.Anything)
}

func TestQuestionService_GetQuestionsByAssessment_RepoError(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	assessmentID := uint(1)
	repoError := errors.New("db find error")

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	mockQuestionRepo.On("FindByAssessmentID", assessmentID).Return(nil, repoError)

	result, err := service.GetQuestionsByAssessment(assessmentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repoError, err)
	mockAssessmentRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}

func TestQuestionService_UpdateQuestion(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	questionID := uint(5)
	existingQuestion := &models.Question{
		ID:            questionID,
		AssessmentID:  1,
		Type:          "multiple-choice",
		Text:          "Old Text",
		CorrectAnswer: "a",
		Points:        5,
		Options: []models.QuestionOption{
			{ID: 10, QuestionID: questionID, OptionID: "a", Text: "Old A"},
			{ID: 11, QuestionID: questionID, OptionID: "b", Text: "Old B"},
		},
	}
	updateData := map[string]interface{}{
		"text":          "New Text",
		"points":        float64(8), // JSON number
		"correctAnswer": "b",
		"options": []interface{}{ // Dữ liệu options từ request thường là []interface{}
			map[string]interface{}{"id": "a", "text": "New A"},
			map[string]interface{}{"id": "b", "text": "New B"},
			map[string]interface{}{"id": "c", "text": "New C"}, // Thêm option mới
		},
	}
	// Question sau khi các trường được cập nhật bởi service
	//updatedQuestionFields := &models.Question{
	//	ID:            questionID,
	//	AssessmentID:  1,
	//	Type:          "multiple-choice",
	//	Text:          "New Text",
	//	CorrectAnswer: "b",
	//	Points:        8,
	//	Options:       existingQuestion.Options, // Options chưa được cập nhật ở bước này
	//}
	// Question cuối cùng trả về sau khi repo xử lý options
	finalUpdatedQuestion := &models.Question{
		ID:            questionID,
		AssessmentID:  1,
		Type:          "multiple-choice",
		Text:          "New Text",
		CorrectAnswer: "b",
		Points:        8,
		Options: []models.QuestionOption{ // Giả sử repo trả về options mới
			{ID: 12, QuestionID: questionID, OptionID: "a", Text: "New A"},
			{ID: 13, QuestionID: questionID, OptionID: "b", Text: "New B"},
			{ID: 14, QuestionID: questionID, OptionID: "c", Text: "New C"},
		},
	}

	mockQuestionRepo.On("FindByID", questionID).Return(existingQuestion, nil).Once()
	// Expect repo Update được gọi với các trường đã cập nhật (trừ options)
	mockQuestionRepo.On("Update", mock.MatchedBy(func(q *models.Question) bool {
		return q.ID == questionID && q.Text == "New Text" && q.Points == 8 && q.CorrectAnswer == "b"
	})).Return(nil)
	// Expect repo DeleteOption được gọi cho các option cũ
	mockQuestionRepo.On("DeleteOption", uint(10)).Return(nil)
	mockQuestionRepo.On("DeleteOption", uint(11)).Return(nil)
	// Expect repo AddOption được gọi cho các option mới
	mockQuestionRepo.On("AddOption", mock.MatchedBy(func(opt *models.QuestionOption) bool { return opt.OptionID == "a" && opt.Text == "New A" })).Return(nil)
	mockQuestionRepo.On("AddOption", mock.MatchedBy(func(opt *models.QuestionOption) bool { return opt.OptionID == "b" && opt.Text == "New B" })).Return(nil)
	mockQuestionRepo.On("AddOption", mock.MatchedBy(func(opt *models.QuestionOption) bool { return opt.OptionID == "c" && opt.Text == "New C" })).Return(nil)
	// Expect repo FindByID được gọi lại cuối cùng để trả về kết quả
	mockQuestionRepo.On("FindByID", questionID).Return(finalUpdatedQuestion, nil).Once()

	result, err := service.UpdateQuestion(questionID, updateData)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "New Text", result.Text)
	assert.Equal(t, 8.0, result.Points)
	assert.Equal(t, "b", result.CorrectAnswer)
	assert.Len(t, result.Options, 3) // Kiểm tra số lượng options mới
	mockQuestionRepo.AssertExpectations(t)
}

func TestQuestionService_UpdateQuestion_NotFound(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	questionID := uint(99)
	updateData := map[string]interface{}{"text": "New Text"}

	mockQuestionRepo.On("FindByID", questionID).Return(nil, errors.New("question not found"))

	result, err := service.UpdateQuestion(questionID, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "question not found")
	mockQuestionRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Update", mock.Anything)
	mockQuestionRepo.AssertNotCalled(t, "DeleteOption", mock.Anything)
	mockQuestionRepo.AssertNotCalled(t, "AddOption", mock.Anything)
}

func TestQuestionService_UpdateQuestion_InvalidCorrectAnswer_MC(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	questionID := uint(5)
	existingQuestion := &models.Question{
		ID: questionID, Type: "multiple-choice",
		Options: []models.QuestionOption{{OptionID: "a"}, {OptionID: "b"}},
	}
	updateData := map[string]interface{}{"correctAnswer": "d"} // 'd' không có trong options

	mockQuestionRepo.On("FindByID", questionID).Return(existingQuestion, nil)

	result, err := service.UpdateQuestion(questionID, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "correct answer must match one of the option IDs")
	mockQuestionRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestQuestionService_UpdateQuestion_InvalidCorrectAnswer_TF(t *testing.T) {
	//mockQuestionRepo := new(MockQuestionRepository)
	//mockAssessmentRepo := new(MockAssessmentRepository)
	//service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)
	//
	//questionID := uint(5)
	//existingQuestion := &models.Question{ID: questionID, Type: "true-false"}
	//updateData := map[string]interface{}{"correctAnswer": "maybe"} // Invalid
	//
	//mockQuestionRepo.On("FindByID", questionID).Return(existingQuestion, nil)
	//
	//result, err := service.UpdateQuestion(questionID, updateData)
	//
	//assert.Error(t, err)
	//assert.Nil(t, result)
	//assert.Contains(t, err.Error(), "correct answer for true-false questions must be a boolean or 'true'/'false' string")
	//mockQuestionRepo.AssertExpectations(t)
	//mockQuestionRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestQuestionService_UpdateQuestion_OptionError(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	questionID := uint(5)
	existingQuestion := &models.Question{
		ID: questionID, Type: "multiple-choice",
		Options: []models.QuestionOption{{ID: 10, OptionID: "a"}},
	}
	updateData := map[string]interface{}{
		"options": []interface{}{map[string]interface{}{"id": "b", "text": "B"}},
	}
	repoError := errors.New("option delete error")

	mockQuestionRepo.On("FindByID", questionID).Return(existingQuestion, nil)
	mockQuestionRepo.On("DeleteOption", uint(10)).Return(repoError) // Giả lập lỗi khi xóa option cũ

	result, err := service.UpdateQuestion(questionID, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repoError, err)
	mockQuestionRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Update", mock.Anything) // Update question không được gọi nếu lỗi option
	mockQuestionRepo.AssertNotCalled(t, "AddOption", mock.Anything)
}

func TestQuestionService_DeleteQuestion(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	questionID := uint(5)
	existingQuestion := &models.Question{ID: questionID}

	mockQuestionRepo.On("FindByID", questionID).Return(existingQuestion, nil)
	mockQuestionRepo.On("Delete", questionID).Return(nil)

	err := service.DeleteQuestion(questionID)

	assert.NoError(t, err)
	mockQuestionRepo.AssertExpectations(t)
}

func TestQuestionService_DeleteQuestion_NotFound(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	questionID := uint(99)
	repoError := errors.New("not found")

	mockQuestionRepo.On("FindByID", questionID).Return(nil, repoError)

	err := service.DeleteQuestion(questionID)

	assert.Error(t, err)
	assert.Equal(t, repoError, err)
	mockQuestionRepo.AssertExpectations(t)
	mockQuestionRepo.AssertNotCalled(t, "Delete", mock.Anything)
}

func TestQuestionService_DeleteQuestion_RepoError(t *testing.T) {
	mockQuestionRepo := new(MockQuestionRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	service := NewQuestionService(mockQuestionRepo, mockAssessmentRepo)

	questionID := uint(5)
	existingQuestion := &models.Question{ID: questionID}
	repoError := errors.New("delete error")

	mockQuestionRepo.On("FindByID", questionID).Return(existingQuestion, nil)
	mockQuestionRepo.On("Delete", questionID).Return(repoError)

	err := service.DeleteQuestion(questionID)

	assert.Error(t, err)
	assert.Equal(t, repoError, err)
	mockQuestionRepo.AssertExpectations(t)
}
