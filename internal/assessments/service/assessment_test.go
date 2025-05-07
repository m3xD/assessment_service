package service

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository implementations
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

// --- Existing Tests ---

func TestCreateAssessment(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	assessment := &models.Assessment{
		Title:        "Test Assessment",
		Subject:      "Math",
		Description:  "Test description",
		Duration:     60,
		CreatedByID:  1,
		PassingScore: 70,
	}

	// Expect Create to be called with an assessment having Status "Draft"
	mockAssessmentRepo.On("Create", mock.MatchedBy(func(a *models.Assessment) bool {
		return a.Title == assessment.Title && a.Status == "Draft"
	})).Return(nil)

	err := service.Create(assessment)

	assert.NoError(t, err)
	assert.Equal(t, "Draft", assessment.Status) // Verify default status
	assert.NotNil(t, assessment.Settings)       // Verify default settings
	// Check default settings values
	assert.False(t, assessment.Settings.RandomizeQuestions)
	assert.True(t, assessment.Settings.ShowResults)
	assert.False(t, assessment.Settings.AllowRetake)
	assert.Equal(t, 1, assessment.Settings.MaxAttempts)
	assert.True(t, assessment.Settings.TimeLimitEnforced)
	// ... add checks for other default settings if needed
	mockAssessmentRepo.AssertExpectations(t)
}

func TestGetByID(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	expectedAssessment := &models.Assessment{
		ID:           1,
		Title:        "Test Assessment",
		Subject:      "Math",
		Description:  "Test description",
		Duration:     60,
		CreatedByID:  1,
		PassingScore: 70,
		Status:       "Active",
	}

	mockAssessmentRepo.On("FindByID", uint(1)).Return(expectedAssessment, nil)

	result, err := service.GetByID(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedAssessment, result)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	mockAssessmentRepo.On("FindByID", uint(99)).Return(nil, errors.New("record not found"))

	result, err := service.GetByID(99)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "record not found")
	mockAssessmentRepo.AssertExpectations(t)
}

func TestUpdateAssessment(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	now := time.Now()
	existingAssessment := &models.Assessment{
		ID:           1,
		Title:        "Old Title",
		Subject:      "Math",
		Description:  "Old description",
		Duration:     60,
		CreatedByID:  1,
		PassingScore: 70,
		Status:       "Draft",
		DueDate:      nil, // Explicitly nil for clarity
	}

	updateData := map[string]interface{}{
		"title":        "New Title",
		"description":  "New description",
		"duration":     float64(90), // JSON numbers are often float64
		"passingScore": float64(80),
		"status":       "Active",
		"dueDate":      now.Format(time.RFC3339),
	}

	mockAssessmentRepo.On("FindByID", uint(1)).Return(existingAssessment, nil)
	// Expect Update to be called with the modified assessment
	mockAssessmentRepo.On("Update", mock.MatchedBy(func(a *models.Assessment) bool {
		return a.ID == 1 &&
			a.Title == "New Title" &&
			a.Description == "New description" &&
			a.Duration == 90 &&
			a.PassingScore == 80 &&
			a.Status == "Active" &&
			a.DueDate != nil && a.DueDate.Format(time.RFC3339) == now.Format(time.RFC3339)
	})).Return(nil)

	result, err := service.Update(1, updateData)

	assert.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
	assert.Equal(t, "New description", result.Description)
	assert.Equal(t, 90, result.Duration)
	assert.Equal(t, float64(80), result.PassingScore)
	assert.Equal(t, "Active", result.Status)
	assert.NotNil(t, result.DueDate)
	assert.Equal(t, now.Format(time.RFC3339), result.DueDate.Format(time.RFC3339))
	mockAssessmentRepo.AssertExpectations(t)
}

func TestUpdateAssessment_NotFound(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	updateData := map[string]interface{}{"title": "New Title"}

	mockAssessmentRepo.On("FindByID", uint(99)).Return(nil, errors.New("record not found"))

	result, err := service.Update(99, updateData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "record not found")
	mockAssessmentRepo.AssertExpectations(t)
	// Ensure Update was NOT called
	mockAssessmentRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestDeleteAssessment(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	existingAssessment := &models.Assessment{ID: 1}

	mockAssessmentRepo.On("FindByID", uint(1)).Return(existingAssessment, nil)
	mockAssessmentRepo.On("Delete", uint(1)).Return(nil)

	err := service.Delete(1)

	assert.NoError(t, err)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestDeleteAssessment_NotFound(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	mockAssessmentRepo.On("FindByID", uint(99)).Return(nil, errors.New("record not found"))

	err := service.Delete(99)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
	mockAssessmentRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertNotCalled(t, "Delete", uint(99))
}

func TestPublishAssessment(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	existingAssessment := &models.Assessment{
		ID:        1,
		Status:    "Draft",
		Questions: []models.Question{{ID: 1, Text: "Test question"}}, // Has questions
	}
	publishedAssessment := &models.Assessment{
		ID:        1,
		Status:    "Active", // Status after publish
		Questions: []models.Question{{ID: 1, Text: "Test question"}},
	}

	mockAssessmentRepo.On("FindByID", uint(1)).Return(existingAssessment, nil).Once() // First call to check
	mockAssessmentRepo.On("Publish", uint(1)).Return(nil)
	mockAssessmentRepo.On("FindByID", uint(1)).Return(publishedAssessment, nil).Once() // Second call to return updated

	result, err := service.Publish(1)

	assert.NoError(t, err)
	assert.Equal(t, "Active", result.Status)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestPublishAssessment_NotFound(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	mockAssessmentRepo.On("FindByID", uint(99)).Return(nil, errors.New("record not found"))

	result, err := service.Publish(99)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "record not found")
	mockAssessmentRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertNotCalled(t, "Publish", uint(99))
}

func TestPublishAssessmentWithoutQuestions(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	existingAssessment := &models.Assessment{
		ID:        1,
		Status:    "Draft",
		Questions: []models.Question{}, // No questions
	}

	mockAssessmentRepo.On("FindByID", uint(1)).Return(existingAssessment, nil)

	_, err := service.Publish(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot publish assessment without questions")
	mockAssessmentRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertNotCalled(t, "Publish", uint(1)) // Publish should not be called
}

// --- New Tests ---

func TestListAssessments(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	params := util.PaginationParams{Page: 0, Limit: 10, Offset: 0}
	expectedAssessments := []models.Assessment{
		{ID: 1, Title: "Assessment 1"},
		{ID: 2, Title: "Assessment 2"},
	}
	expectedTotal := int64(5) // Example total count

	mockAssessmentRepo.On("List", params).Return(expectedAssessments, expectedTotal, nil)

	assessments, total, err := service.List(params)

	assert.NoError(t, err)
	assert.Equal(t, expectedAssessments, assessments)
	assert.Equal(t, expectedTotal, total)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestGetRecentAssessments(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	limit := 5
	expectedAssessments := []models.Assessment{
		{ID: 5, Title: "Recent 5"},
		{ID: 4, Title: "Recent 4"},
	}

	mockAssessmentRepo.On("FindRecent", limit).Return(expectedAssessments, nil)

	assessments, err := service.GetRecentAssessments(limit)

	assert.NoError(t, err)
	assert.Equal(t, expectedAssessments, assessments)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestGetStatistics(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	expectedStats := map[string]interface{}{
		"totalAssessments":   int64(10),
		"activeAssessments":  int64(5),
		"draftAssessments":   int64(3),
		"expiredAssessments": int64(2),
		"totalAttempts":      int64(150),
		"passRate":           75.5,
		"averageScore":       82.1,
		"bySubject": []map[string]interface{}{
			{"subject": "Math", "count": 4},
			{"subject": "Science", "count": 6},
		},
	}

	mockAssessmentRepo.On("GetStatistics").Return(expectedStats, nil)

	stats, err := service.GetStatistics()

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestUpdateSettings(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	assessmentID := uint(1)
	existingAssessment := &models.Assessment{ID: assessmentID} // Need something for FindByID
	newSettings := &models.AssessmentSettings{
		RandomizeQuestions: true,
		ShowResults:        false,
		MaxAttempts:        3,
	}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(existingAssessment, nil)
	mockAssessmentRepo.On("UpdateSettings", assessmentID, newSettings).Return(nil)

	err := service.UpdateSettings(assessmentID, newSettings)

	assert.NoError(t, err)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestUpdateSettings_AssessmentNotFound(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	assessmentID := uint(99)
	newSettings := &models.AssessmentSettings{}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(nil, errors.New("record not found"))

	err := service.UpdateSettings(assessmentID, newSettings)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
	mockAssessmentRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertNotCalled(t, "UpdateSettings", assessmentID, newSettings)
}

func TestGetResults(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	assessmentID := uint(1)
	params := util.PaginationParams{Page: 0, Limit: 10}
	expectedResults := []map[string]interface{}{
		{"user": "Alice", "score": 90.0},
		{"user": "Bob", "score": 75.0},
	}
	expectedTotal := int64(20)

	mockAssessmentRepo.On("GetResults", assessmentID, params).Return(expectedResults, expectedTotal, nil)

	results, total, err := service.GetResults(assessmentID, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)
	assert.Equal(t, expectedTotal, total)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestDuplicateAssessment(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	originalID := uint(1)
	originalAssessment := &models.Assessment{
		ID:           originalID,
		Title:        "Original Title",
		Subject:      "Science",
		Description:  "Original Desc",
		Duration:     30,
		CreatedByID:  2,
		PassingScore: 65,
		Status:       "Active",
		Questions:    []models.Question{{ID: 10, Text: "Q1"}},
		Settings:     models.AssessmentSettings{ID: 1, RandomizeQuestions: true},
	}

	// Define the assessment state *after* modification before calling Duplicate repo method
	expectedDuplicateInput := &models.Assessment{
		// ID will be zero before creation by Duplicate
		Title:        "New Copy Title", // New title provided
		Subject:      "Science",
		Description:  "Original Desc",
		Duration:     30,
		CreatedByID:  2,
		PassingScore: 65,
		Status:       "draft",                                                    // setAsDraft is true
		Questions:    []models.Question{{ID: 10, Text: "Q1"}},                    // copyQuestions is true
		Settings:     models.AssessmentSettings{ID: 1, RandomizeQuestions: true}, // copySettings is true
	}

	// Define the assessment returned by the final List call
	duplicatedAssessmentResult := models.Assessment{
		ID:           2, // New ID after creation
		Title:        "New Copy Title",
		Subject:      "Science",
		Description:  "Original Desc",
		Duration:     30,
		CreatedByID:  2,
		PassingScore: 65,
		Status:       "draft",
		// Questions and Settings might or might not be loaded by List depending on repo implementation
	}

	mockAssessmentRepo.On("FindByID", originalID).Return(originalAssessment, nil).Once()
	mockAssessmentRepo.On("Duplicate", mock.MatchedBy(func(a *models.Assessment) bool {
		// Check the state passed to the Duplicate repository method
		return a.Title == expectedDuplicateInput.Title &&
			a.Subject == expectedDuplicateInput.Subject &&
			a.Status == expectedDuplicateInput.Status &&
			len(a.Questions) == len(expectedDuplicateInput.Questions) && // Check questions are copied
			a.Settings.RandomizeQuestions == expectedDuplicateInput.Settings.RandomizeQuestions // Check settings are copied
	})).Return(nil)

	// Mock the List call used to retrieve the newly created assessment
	listParams := util.PaginationParams{
		Limit: 1,
		Filters: map[string]interface{}{
			"title": "New Copy Title",
		},
		SortBy:  "created_at",
		SortDir: "DESC",
	}
	mockAssessmentRepo.On("List", listParams).Return([]models.Assessment{duplicatedAssessmentResult}, int64(1), nil)

	// Call the service method
	newTitle := "New Copy Title"
	copyQuestions := true
	copySettings := true
	setAsDraft := true
	result, err := service.Duplicate(originalID, newTitle, copyQuestions, copySettings, setAsDraft)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, duplicatedAssessmentResult.ID, result.ID) // Check if the correct new assessment is returned
	assert.Equal(t, newTitle, result.Title)
	assert.Equal(t, "draft", result.Status)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestDuplicateAssessment_Defaults(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	originalID := uint(1)
	originalAssessment := &models.Assessment{
		ID:           originalID,
		Title:        "Original Title",
		Subject:      "Science",
		Description:  "Original Desc",
		Duration:     30,
		CreatedByID:  2,
		PassingScore: 65,
		Status:       "Active", // Original status
		Questions:    []models.Question{{ID: 10, Text: "Q1"}},
		Settings:     models.AssessmentSettings{ID: 1, RandomizeQuestions: true},
	}

	// Define the assessment state *after* modification before calling Duplicate repo method
	expectedDuplicateInput := &models.Assessment{
		Title:        "Original Title (Copy)", // Default title
		Subject:      "Science",
		Description:  "Original Desc",
		Duration:     30,
		CreatedByID:  2,
		PassingScore: 65,
		Status:       "Active",                    // Default status (not setAsDraft)
		Questions:    []models.Question{},         // Default: Don't copy questions
		Settings:     models.AssessmentSettings{}, // Default: Don't copy settings
	}

	// Define the assessment returned by the final List call
	duplicatedAssessmentResult := models.Assessment{
		ID:    2,
		Title: "Original Title (Copy)",
		// ... other fields
	}

	mockAssessmentRepo.On("FindByID", originalID).Return(originalAssessment, nil).Once()
	mockAssessmentRepo.On("Duplicate", mock.MatchedBy(func(a *models.Assessment) bool {
		return a.Title == expectedDuplicateInput.Title &&
			a.Status == expectedDuplicateInput.Status &&
			len(a.Questions) == 0 && // Check questions are NOT copied
			a.Settings.ID == 0 // Check settings are NOT copied (assuming ID is zero for empty)
	})).Return(nil)

	// Mock the List call
	listParams := util.PaginationParams{
		Limit: 1,
		Filters: map[string]interface{}{
			"title": "Original Title (Copy)",
		},
		SortBy:  "created_at",
		SortDir: "DESC",
	}
	mockAssessmentRepo.On("List", listParams).Return([]models.Assessment{duplicatedAssessmentResult}, int64(1), nil)

	// Call the service method with defaults
	result, err := service.Duplicate(originalID, "", false, false, false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, duplicatedAssessmentResult.ID, result.ID)
	assert.Equal(t, "Original Title (Copy)", result.Title)
	// Status should match the original if setAsDraft is false
	// assert.Equal(t, originalAssessment.Status, result.Status)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestGetAssessmentHasAttempt(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	userID := uint(1)
	params := util.PaginationParams{Page: 0, Limit: 5}
	expectedAssessments := []models.Assessment{
		{ID: 10, Title: "Attempted Assessment 1"},
		{ID: 12, Title: "Attempted Assessment 2"},
	}
	expectedTotal := int64(2)

	mockAssessmentRepo.On("GetAssessmentHasAttemptByUser", params, userID).Return(expectedAssessments, expectedTotal, nil)

	assessments, total, err := service.GetAssessmentHasAttempt(userID, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedAssessments, assessments)
	assert.Equal(t, expectedTotal, total)
	mockAssessmentRepo.AssertExpectations(t)
}

func TestGetAssessmentDetailWithUser(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	assessmentID := uint(1)
	params := util.PaginationParams{Page: 0, Limit: 10}
	expectedAssessment := &models.Assessment{ID: assessmentID, Title: "Assessment Detail"}
	expectedUsers := []models.User{
		{ID: 1, Name: "User A"},
		{ID: 2, Name: "User B"},
	}
	expectedTotalUsers := int64(5)

	mockAssessmentRepo.On("FindByID", assessmentID).Return(expectedAssessment, nil)
	mockUserRepo.On("GetListUserByAssessment", params, assessmentID).Return(expectedUsers, expectedTotalUsers, nil)

	assessment, users, total, err := service.GetAssessmentDetailWithUser(assessmentID, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedAssessment, assessment)
	assert.Equal(t, expectedUsers, users)
	assert.Equal(t, expectedTotalUsers, total)
	mockAssessmentRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGetAssessmentDetailWithUser_AssessmentNotFound(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewAssessmentService(mockAssessmentRepo, mockUserRepo)

	assessmentID := uint(99)
	params := util.PaginationParams{}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(nil, errors.New("record not found"))

	assessment, users, total, err := service.GetAssessmentDetailWithUser(assessmentID, params)

	assert.Error(t, err)
	assert.Nil(t, assessment)
	assert.Nil(t, users)
	assert.Zero(t, total)
	assert.Contains(t, err.Error(), "record not found")
	mockAssessmentRepo.AssertExpectations(t)
	// Ensure user repo method is not called if assessment is not found
	mockUserRepo.AssertNotCalled(t, "GetListUserByAssessment", mock.Anything, mock.Anything)
}
