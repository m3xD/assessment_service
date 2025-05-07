package rest

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// --- Mock AssessmentService ---
type MockAssessmentService struct {
	mock.Mock
}

// Implement AssessmentService interface for mock
func (m *MockAssessmentService) Create(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}

func (m *MockAssessmentService) GetByID(id uint) (*models.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	assessment, ok := args.Get(0).(*models.Assessment)
	if !ok && args.Get(0) != nil {
		panic("Mock GetByID returned non-nil value of incorrect type")
	}
	return assessment, args.Error(1)
}

func (m *MockAssessmentService) Update(id uint, assessmentData map[string]interface{}) (*models.Assessment, error) {
	args := m.Called(id, assessmentData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	assessment, ok := args.Get(0).(*models.Assessment)
	if !ok && args.Get(0) != nil {
		panic("Mock Update returned non-nil value of incorrect type")
	}
	return assessment, args.Error(1)
}

func (m *MockAssessmentService) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAssessmentService) List(params util.PaginationParams) ([]models.Assessment, int64, error) {
	args := m.Called(params)
	assessments, ok := args.Get(0).([]models.Assessment)
	if !ok && args.Get(0) != nil {
		// Return empty slice if type assertion fails but value is not nil
		assessments = []models.Assessment{}
	}
	count, ok := args.Get(1).(int64)
	if !ok {
		if intCount, okInt := args.Get(1).(int); okInt {
			count = int64(intCount)
		} else {
			count = 0 // Default to 0 if type assertion fails
		}
	}
	return assessments, count, args.Error(2)
}

func (m *MockAssessmentService) GetRecentAssessments(limit int) ([]models.Assessment, error) {
	args := m.Called(limit)
	assessments, ok := args.Get(0).([]models.Assessment)
	if !ok && args.Get(0) != nil {
		// Return empty slice if type assertion fails but value is not nil
		assessments = []models.Assessment{}
	}
	return assessments, args.Error(1)
}

func (m *MockAssessmentService) GetStatistics() (map[string]interface{}, error) {
	args := m.Called()
	stats, ok := args.Get(0).(map[string]interface{})
	if !ok && args.Get(0) != nil {
		panic("Mock GetStatistics returned non-nil value of incorrect type")
	}
	return stats, args.Error(1)
}

func (m *MockAssessmentService) UpdateSettings(id uint, settings *models.AssessmentSettings) error {
	args := m.Called(id, settings)
	return args.Error(0)
}

func (m *MockAssessmentService) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(id, params)
	results, ok := args.Get(0).([]map[string]interface{})
	if !ok && args.Get(0) != nil {
		// Return empty slice if type assertion fails but value is not nil
		results = []map[string]interface{}{}
	}
	count, ok := args.Get(1).(int64)
	if !ok {
		if intCount, okInt := args.Get(1).(int); okInt {
			count = int64(intCount)
		} else {
			count = 0 // Default to 0 if type assertion fails
		}
	}
	return results, count, args.Error(2)
}

func (m *MockAssessmentService) Publish(id uint) (*models.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	assessment, ok := args.Get(0).(*models.Assessment)
	if !ok && args.Get(0) != nil {
		panic("Mock Publish returned non-nil value of incorrect type")
	}
	return assessment, args.Error(1)
}

func (m *MockAssessmentService) Duplicate(id uint, newTitle string, copyQuestions, copySettings, setAsDraft bool) (*models.Assessment, error) {
	args := m.Called(id, newTitle, copyQuestions, copySettings, setAsDraft)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	assessment, ok := args.Get(0).(*models.Assessment)
	if !ok && args.Get(0) != nil {
		panic("Mock Duplicate returned non-nil value of incorrect type")
	}
	return assessment, args.Error(1)
}

func (m *MockAssessmentService) GetAssessmentDetailWithUser(assessmentID uint, params util.PaginationParams) (*models.Assessment, []models.User, int64, error) {
	args := m.Called(assessmentID, params)
	assessment, _ := args.Get(0).(*models.Assessment)
	users, _ := args.Get(1).([]models.User)
	total, _ := args.Get(2).(int64)
	return assessment, users, total, args.Error(3)
}

func (m *MockAssessmentService) GetAssessmentHasAttempt(userID uint, params util.PaginationParams) ([]models.Assessment, int64, error) {
	args := m.Called(userID, params)
	assessments, _ := args.Get(0).([]models.Assessment)
	total, _ := args.Get(1).(int64)
	return assessments, total, args.Error(2)
}

// Helper function to create a request with context containing JWT claims
func createRequestWithClaims(method, url string, body []byte, claims jwt.MapClaims) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	if claims != nil {
		ctx := context.WithValue(req.Context(), "user", claims)
		req = req.WithContext(ctx)
	}
	return req
}

// --- Test Cases ---

func TestAssessmentHandler_CreateAssessment(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	createReq := map[string]interface{}{
		"title":        "New Assessment",
		"subject":      "Testing",
		"description":  "A test assessment",
		"duration":     60,
		"dueDate":      time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"passingScore": 75.0, // Use float64 for JSON numbers
		"status":       "Draft",
	}
	body, _ := json.Marshal(createReq)

	mockService.On("Create", mock.MatchedBy(func(a *models.Assessment) bool {
		return a.Title == createReq["title"] &&
			a.Subject == createReq["subject"] &&
			a.CreatedByID == uint(123) &&
			a.Status == createReq["status"]
	})).Return(nil).Run(func(args mock.Arguments) {
		assessment := args.Get(0).(*models.Assessment)
		assessment.ID = 1
	})

	claims := jwt.MapClaims{"userID": "123"}
	req := createRequestWithClaims(http.MethodPost, "/assessments", body, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments", handler.CreateAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertExpectations(t)

	var responseAssessment models.Assessment
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssessment)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), responseAssessment.ID)
	assert.Equal(t, createReq["title"], responseAssessment.Title)
}

func TestAssessmentHandler_CreateAssessment_InvalidInput(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	invalidBody := []byte(`{"passingScore": "Missing Subject"}`)

	claims := jwt.MapClaims{"userID": "123"}
	req := createRequestWithClaims(http.MethodPost, "/assessments", invalidBody, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments", handler.CreateAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "Create", mock.Anything)
}

func TestAssessmentHandler_CreateAssessment_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	createReq := map[string]interface{}{
		"title":        "New Assessment",
		"subject":      "Testing",
		"duration":     60,
		"passingScore": 75.0,
	}
	body, _ := json.Marshal(createReq)

	mockService.On("Create", mock.Anything).Return(errors.New("database error"))

	claims := jwt.MapClaims{"userID": "123"}
	req := createRequestWithClaims(http.MethodPost, "/assessments", body, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments", handler.CreateAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_GetAssessmentById(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(5)
	expectedAssessment := &models.Assessment{
		ID:    assessmentID,
		Title: "Found Assessment",
	}

	mockService.On("GetByID", assessmentID).Return(expectedAssessment, nil)

	req := httptest.NewRequest(http.MethodGet, "/assessments/"+strconv.Itoa(int(assessmentID)), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}", handler.GetAssessmentById).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var responseAssessment models.Assessment
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssessment)
	assert.NoError(t, err)
	assert.Equal(t, expectedAssessment.ID, responseAssessment.ID)
	assert.Equal(t, expectedAssessment.Title, responseAssessment.Title)
}

func TestAssessmentHandler_GetAssessmentById_NotFound(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(999)

	mockService.On("GetByID", assessmentID).Return(nil, errors.New("assessment not found")) // Simulate not found

	req := httptest.NewRequest(http.MethodGet, "/assessments/"+strconv.Itoa(int(assessmentID)), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}", handler.GetAssessmentById).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_GetAssessmentById_InvalidID(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/assessments/invalid-id", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	// Test the handler's internal parsing, even if the route regex catches it
	router.HandleFunc("/assessments/{id}", handler.GetAssessmentById).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetByID", mock.Anything)
}

func TestAssessmentHandler_UpdateAssessment(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	updateReq := map[string]interface{}{
		"title":       "Updated Title",
		"description": "Updated Desc",
		"duration":    90.0, // Use float64
	}
	body, _ := json.Marshal(updateReq)

	updatedAssessment := &models.Assessment{
		ID:          assessmentID,
		Title:       "Updated Title",
		Description: "Updated Desc",
		Duration:    90,
		// ... other fields remain unchanged from original fetch in service
	}

	// Expect Update to be called with the correct ID and data map
	mockService.On("Update", assessmentID, mock.Anything).Return(updatedAssessment, nil)

	req := httptest.NewRequest(http.MethodPut, "/assessments/"+strconv.Itoa(int(assessmentID)), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}", handler.UpdateAssessment).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var responseAssessment models.Assessment
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssessment)
	assert.NoError(t, err)
	assert.Equal(t, updatedAssessment.ID, responseAssessment.ID)
	assert.Equal(t, updatedAssessment.Title, responseAssessment.Title)
	assert.Equal(t, updatedAssessment.Duration, responseAssessment.Duration)
}

func TestAssessmentHandler_UpdateAssessment_InvalidInput(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	invalidBody := []byte(`{"duration": "not-a-number"}`)

	req := httptest.NewRequest(http.MethodPut, "/assessments/"+strconv.Itoa(int(assessmentID)), bytes.NewBuffer(invalidBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}", handler.UpdateAssessment).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestAssessmentHandler_UpdateAssessment_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	updateReq := map[string]interface{}{"title": "Updated Title"}
	body, _ := json.Marshal(updateReq)

	mockService.On("Update", assessmentID, mock.Anything).Return(nil, errors.New("update failed"))

	req := httptest.NewRequest(http.MethodPut, "/assessments/"+strconv.Itoa(int(assessmentID)), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}", handler.UpdateAssessment).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_DeleteAssessment(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)

	mockService.On("Delete", assessmentID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/assessments/"+strconv.Itoa(int(assessmentID)), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}", handler.DeleteAssessment).Methods(http.MethodDelete)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	// Check response message
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "OK", resp["status"])
	assert.Equal(t, "Assessment deleted successfully", resp["message"])
}

func TestAssessmentHandler_DeleteAssessment_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)

	mockService.On("Delete", assessmentID).Return(errors.New("deletion failed"))

	req := httptest.NewRequest(http.MethodDelete, "/assessments/"+strconv.Itoa(int(assessmentID)), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}", handler.DeleteAssessment).Methods(http.MethodDelete)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_ListAssessments(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	expectedAssessments := []models.Assessment{
		{ID: 1, Title: "Assessment 1", Subject: "Math", Status: "Active"},
		{ID: 2, Title: "Assessment 2", Subject: "Science", Status: "Draft"},
	}
	expectedTotal := int64(5)

	// Define expected params based on request URL
	expectedParams := util.PaginationParams{
		Page:    1,
		Limit:   5,
		Offset:  5, // page * limit
		SortBy:  "title",
		SortDir: "ASC",
		Filters: map[string]interface{}{"subject": "Math", "status": "Active"},
	}

	mockService.On("List", expectedParams).Return(expectedAssessments, expectedTotal, nil)

	req := httptest.NewRequest(http.MethodGet, "/assessments?page=1&pageSize=5&sort=title,asc&subject=Math&status=Active", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments", handler.ListAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	// Assert pagination response structure
	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(expectedTotal), result["totalElements"]) // JSON numbers are float64
	assert.Equal(t, float64(expectedParams.Page), result["number"])
	content, ok := result["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, content, len(expectedAssessments))
}

func TestAssessmentHandler_ListAssessments_Empty(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	expectedAssessments := []models.Assessment{} // Empty list
	expectedTotal := int64(0)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	mockService.On("List", expectedParams).Return(expectedAssessments, expectedTotal, nil)

	req := httptest.NewRequest(http.MethodGet, "/assessments", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments", handler.ListAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), result["totalElements"])
	content, ok := result["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, content, 0)
	assert.True(t, result["empty"].(bool))
}

func TestAssessmentHandler_ListAssessments_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	mockService.On("List", mock.Anything).Return(nil, int64(0), errors.New("list error"))

	req := httptest.NewRequest(http.MethodGet, "/assessments", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments", handler.ListAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_GetRecentAssessments(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	expectedLimit := 7
	expectedAssessments := []models.Assessment{
		{ID: 10, Title: "Recent 1"},
		{ID: 9, Title: "Recent 2"},
	}

	mockService.On("GetRecentAssessments", expectedLimit).Return(expectedAssessments, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/assessments/recent?limit=%d", expectedLimit), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/recent", handler.GetRecentAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var responseAssessments []models.Assessment
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssessments)
	assert.NoError(t, err)
	assert.Equal(t, expectedAssessments, responseAssessments)
}

func TestAssessmentHandler_GetRecentAssessments_InvalidLimit(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	defaultLimit := 5 // Default limit when parsing fails
	expectedAssessments := []models.Assessment{{ID: 1, Title: "Default Limit Result"}}

	mockService.On("GetRecentAssessments", defaultLimit).Return(expectedAssessments, nil)

	req := httptest.NewRequest(http.MethodGet, "/assessments/recent?limit=invalid", nil) // Invalid limit value
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/recent", handler.GetRecentAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code) // Should still succeed with default limit
	mockService.AssertExpectations(t)       // Verify service was called with default limit

	var responseAssessments []models.Assessment
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssessments)
	assert.NoError(t, err)
	assert.Equal(t, expectedAssessments, responseAssessments)
}

func TestAssessmentHandler_GetRecentAssessments_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	expectedLimit := 5

	mockService.On("GetRecentAssessments", expectedLimit).Return(nil, errors.New("fetch error"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/assessments/recent?limit=%d", expectedLimit), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/recent", handler.GetRecentAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_GetAssessmentStatistics(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	expectedStats := map[string]interface{}{"totalAssessments": 10.0} // Use float64 for JSON numbers

	mockService.On("GetStatistics").Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/assessments/statistics", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/statistics", handler.GetAssessmentStatistics).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var responseStats map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &responseStats)
	assert.NoError(t, err)
	assert.Equal(t, expectedStats, responseStats)
}

func TestAssessmentHandler_GetAssessmentStatistics_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	mockService.On("GetStatistics").Return(nil, errors.New("stats error"))

	req := httptest.NewRequest(http.MethodGet, "/assessments/statistics", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/statistics", handler.GetAssessmentStatistics).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_UpdateSettings(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	settingsReq := models.AssessmentSettings{
		RandomizeQuestions: true,
		ShowResults:        false,
		MaxAttempts:        2,
	}
	body, _ := json.Marshal(settingsReq)

	mockService.On("UpdateSettings", assessmentID, &settingsReq).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/assessments/"+strconv.Itoa(int(assessmentID))+"/settings", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/settings", handler.UpdateSettings).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	// Check response structure
	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(assessmentID), resp["id"]) // JSON number
	settingsResp, ok := resp["settings"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, settingsReq.RandomizeQuestions, settingsResp["randomizeQuestions"])
	assert.Equal(t, settingsReq.ShowResults, settingsResp["showResults"])
}

func TestAssessmentHandler_UpdateSettings_InvalidInput(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	invalidBody := []byte(`{"maxAttempts": "not-a-number"}`)

	req := httptest.NewRequest(http.MethodPut, "/assessments/"+strconv.Itoa(int(assessmentID))+"/settings", bytes.NewBuffer(invalidBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/settings", handler.UpdateSettings).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "UpdateSettings", mock.Anything, mock.Anything)
}

func TestAssessmentHandler_UpdateSettings_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	settingsReq := models.AssessmentSettings{MaxAttempts: 3}
	body, _ := json.Marshal(settingsReq)

	mockService.On("UpdateSettings", assessmentID, &settingsReq).Return(errors.New("update settings failed"))

	req := httptest.NewRequest(http.MethodPut, "/assessments/"+strconv.Itoa(int(assessmentID))+"/settings", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/settings", handler.UpdateSettings).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_GetAssessmentResults(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	expectedResults := []map[string]interface{}{
		{"user": "User1", "score": 80.0},
	}
	expectedTotal := int64(15)
	expectedParams := util.PaginationParams{
		Page:    0,
		Limit:   10,
		Offset:  0,
		Filters: map[string]interface{}{"user": "User1"},
		// Default sort applied by GetPaginationParams if not in URL
		SortBy:  "created_at",
		SortDir: "DESC",
	}

	mockService.On("GetResults", assessmentID, expectedParams).Return(expectedResults, expectedTotal, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/assessments/%d/results?user=User1", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/results", handler.GetAssessmentResults).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	// Assert pagination response structure
	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(expectedTotal), result["totalElements"])
	content, ok := result["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, content, len(expectedResults))
}

func TestAssessmentHandler_GetAssessmentResults_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	mockService.On("GetResults", assessmentID, expectedParams).Return(nil, int64(0), errors.New("results error"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/assessments/%d/results", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/results", handler.GetAssessmentResults).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_PublishAssessment(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	publishedAssessment := &models.Assessment{
		ID:     assessmentID,
		Status: "Active",
		Title:  "Published Assessment",
	}

	mockService.On("Publish", assessmentID).Return(publishedAssessment, nil)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/publish", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/publish", handler.PublishAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var responseAssessment models.Assessment
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssessment)
	assert.NoError(t, err)
	assert.Equal(t, publishedAssessment.ID, responseAssessment.ID)
	assert.Equal(t, publishedAssessment.Status, responseAssessment.Status)
}

func TestAssessmentHandler_PublishAssessment_NoQuestionsError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	noQuestionsError := errors.New("cannot publish assessment without questions")

	mockService.On("Publish", assessmentID).Return(nil, noQuestionsError)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/publish", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/publish", handler.PublishAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code) // Or potentially BadRequest depending on desired behavior
	mockService.AssertExpectations(t)
	// Check error message in response
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["message"], "Failed to publish assessment") // Generic message is okay
}

func TestAssessmentHandler_PublishAssessment_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)

	mockService.On("Publish", assessmentID).Return(nil, errors.New("publish error"))

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/publish", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/publish", handler.PublishAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_DuplicateAssessment(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	originalID := uint(1)
	duplicateReq := map[string]interface{}{
		"newTitle":      "Duplicated Title",
		"copyQuestions": true,
		"copySettings":  false,
		"setAsDraft":    true,
	}
	body, _ := json.Marshal(duplicateReq)

	duplicatedAssessment := &models.Assessment{
		ID:    2, // New ID
		Title: "Duplicated Title",
		// ... other fields
	}

	mockService.On("Duplicate", originalID, "Duplicated Title", true, false, true).Return(duplicatedAssessment, nil)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/duplicate", originalID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/duplicate", handler.DuplicateAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertExpectations(t)

	var responseAssessment models.Assessment
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssessment)
	assert.NoError(t, err)
	assert.Equal(t, duplicatedAssessment.ID, responseAssessment.ID)
	assert.Equal(t, duplicatedAssessment.Title, responseAssessment.Title)
}

func TestAssessmentHandler_DuplicateAssessment_InvalidInput(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	originalID := uint(1)
	invalidBody := []byte(`{"copyQuestions": "not-a-boolean"}`)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/duplicate", originalID), bytes.NewBuffer(invalidBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/duplicate", handler.DuplicateAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "Duplicate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAssessmentHandler_DuplicateAssessment_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	originalID := uint(1)
	duplicateReq := map[string]interface{}{"newTitle": "Duplicated Title"}
	body, _ := json.Marshal(duplicateReq)

	mockService.On("Duplicate", originalID, "Duplicated Title", false, false, false).Return(nil, errors.New("duplicate error")) // Assuming defaults for bools if not provided

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/duplicate", originalID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/duplicate", handler.DuplicateAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_GetAssessmentWithUserHasAttempt(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	expectedAssessment := &models.Assessment{ID: assessmentID, Title: "Detail Assessment"}
	expectedUsers := []models.User{{ID: 1, Name: "User A"}}
	expectedTotalUsers := int64(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	mockService.On("GetAssessmentDetailWithUser", assessmentID, expectedParams).Return(expectedAssessment, expectedUsers, expectedTotalUsers, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/assessments/%d", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	// Assuming the route is defined like this in your actual router setup
	router.HandleFunc("/admin/assessments/{assessmentID:[0-9]+}", handler.GetAssessmentWithUserHasAttempt).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check assessment part
	assessmentResp, ok := resp["assessment"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(expectedAssessment.ID), assessmentResp["id"])
	assert.Equal(t, expectedAssessment.Title, assessmentResp["title"])

	// Check users pagination part
	usersResp, ok := resp["users"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(expectedTotalUsers), usersResp["totalElements"])
	usersContent, ok := usersResp["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, usersContent, len(expectedUsers))
}

func TestAssessmentHandler_GetAssessmentWithUserHasAttempt_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	assessmentID := uint(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	mockService.On("GetAssessmentDetailWithUser", assessmentID, expectedParams).Return(nil, nil, int64(0), errors.New("fetch error"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/assessments/%d", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/assessments/{assessmentID:[0-9]+}", handler.GetAssessmentWithUserHasAttempt).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAssessmentHandler_GetAssessmentHasBeenAttemptByUser(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	userID := uint(7)
	expectedAssessments := []models.Assessment{{ID: 1, Title: "Attempted 1"}}
	expectedTotal := int64(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	mockService.On("GetAssessmentHasAttempt", userID, expectedParams).Return(expectedAssessments, expectedTotal, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/assessments/attempted/%d", userID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/assessments/attempted/{userID:[0-9]+}", handler.GetAssessmentHasBeenAttemptByUser).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	// Assert pagination response structure
	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(expectedTotal), result["totalElements"])
	content, ok := result["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, content, len(expectedAssessments))
}

func TestAssessmentHandler_GetAssessmentHasBeenAttemptByUser_InvalidUserID(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/assessments/attempted/invalid", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/assessments/attempted/{userID}", handler.GetAssessmentHasBeenAttemptByUser).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetAssessmentHasAttempt", mock.Anything, mock.Anything)
}

func TestAssessmentHandler_GetAssessmentHasBeenAttemptByUser_ServiceError(t *testing.T) {
	mockService := new(MockAssessmentService)
	logger := zaptest.NewLogger(t)
	handler := NewAssessmentHandler(mockService, logger)

	userID := uint(7)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	mockService.On("GetAssessmentHasAttempt", userID, expectedParams).Return(nil, int64(0), errors.New("fetch error"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/assessments/attempted/%d", userID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/assessments/attempted/{userID:[0-9]+}", handler.GetAssessmentHasBeenAttemptByUser).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}
