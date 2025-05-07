package delivery

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// --- Mock AttemptService ---
type MockAttemptService struct {
	mock.Mock
}

func (m *MockAttemptService) GetListAttemptByUserAndAssessment(userID, assessmentID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, assessmentID, params)
	attempts, _ := args.Get(0).([]models.Attempt)
	total, _ := args.Get(1).(int64)
	return attempts, total, args.Error(2)
}

func (m *MockAttemptService) GetAttemptDetail(attemptID uint) (*models.Attempt, error) {
	args := m.Called(attemptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	attempt, ok := args.Get(0).(*models.Attempt)
	if !ok && args.Get(0) != nil {
		panic("Mock GetAttemptDetail returned non-nil value of incorrect type")
	}
	return attempt, args.Error(1)
}

func (m *MockAttemptService) GradeAttempt(newAttempt models.AttemptUpdateDTO, attemptID uint) error {
	args := m.Called(newAttempt, attemptID)
	return args.Error(0)
}

// --- Test Cases ---

func TestAttemptHandler_GetListAttemptByUserAndAssessment(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	userID := uint(1)
	assessmentID := uint(10)
	expectedAttempts := []models.Attempt{{ID: 1, UserID: userID, AssessmentID: assessmentID}}
	expectedTotal := int64(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	mockService.On("GetListAttemptByUserAndAssessment", userID, assessmentID, expectedParams).Return(expectedAttempts, expectedTotal, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/attempts/%d/users/%d", assessmentID, userID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempts/{assessmentID:[0-9]+}/users/{userID:[0-9]+}", handler.GetListAttemptByUserAndAssessment).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(expectedTotal), result["totalElements"])
	content, ok := result["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, content, len(expectedAttempts))
}

func TestAttemptHandler_GetListAttemptByUserAndAssessment_InvalidUserID(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/attempts/10/users/invalid", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempts/{assessmentID:[0-9]+}/users/{userID}", handler.GetListAttemptByUserAndAssessment).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetListAttemptByUserAndAssessment", mock.Anything, mock.Anything, mock.Anything)
}

func TestAttemptHandler_GetListAttemptByUserAndAssessment_InvalidAssessmentID(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/attempts/invalid/users/1", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempts/{assessmentID}/users/{userID:[0-9]+}", handler.GetListAttemptByUserAndAssessment).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetListAttemptByUserAndAssessment", mock.Anything, mock.Anything, mock.Anything)
}

func TestAttemptHandler_GetListAttemptByUserAndAssessment_ServiceError(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	userID := uint(1)
	assessmentID := uint(10)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}}

	mockService.On("GetListAttemptByUserAndAssessment", userID, assessmentID, expectedParams).Return(nil, int64(0), errors.New("service error"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/attempts/%d/users/%d", assessmentID, userID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempts/{assessmentID:[0-9]+}/users/{userID:[0-9]+}", handler.GetListAttemptByUserAndAssessment).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAttemptHandler_GetAttemptDetail(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	attemptID := uint(5)
	expectedAttempt := &models.Attempt{ID: attemptID, Status: "Completed"}

	mockService.On("GetAttemptDetail", attemptID).Return(expectedAttempt, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/attempt/%d/users/1", attemptID), nil) // UserID không quan trọng trong test này
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempt/{attemptID:[0-9]+}/users/{userID:[0-9]+}", handler.GetAttemptDetail).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var responseAttempt models.Attempt
	err := json.Unmarshal(rr.Body.Bytes(), &responseAttempt)
	assert.NoError(t, err)
	assert.Equal(t, expectedAttempt.ID, responseAttempt.ID)
	assert.Equal(t, expectedAttempt.Status, responseAttempt.Status)
}

func TestAttemptHandler_GetAttemptDetail_InvalidAttemptID(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/attempt/invalid/users/1", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempt/{attemptID}/users/{userID:[0-9]+}", handler.GetAttemptDetail).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetAttemptDetail", mock.Anything)
}

func TestAttemptHandler_GetAttemptDetail_ServiceError(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	attemptID := uint(99)

	mockService.On("GetAttemptDetail", attemptID).Return(nil, errors.New("fetch error"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/attempt/%d/users/1", attemptID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempt/{attemptID:[0-9]+}/users/{userID:[0-9]+}", handler.GetAttemptDetail).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAttemptHandler_GradeAttempt(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	attemptID := uint(1)
	gradeReq := models.AttemptUpdateDTO{
		Score:    95.0,
		Feedback: "Excellent work",
		Answers: []struct {
			ID        uint `json:"id"`
			IsCorrect bool `json:"isCorrect"`
		}{{ID: 10, IsCorrect: true}},
	}
	body, _ := json.Marshal(gradeReq)

	mockService.On("GradeAttempt", gradeReq, attemptID).Return(nil)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/admin/attempt/grade/%d", attemptID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempt/grade/{attemptID:[0-9]+}", handler.GradeAttempt).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "OK", resp["status"])
	assert.Equal(t, "Successfully graded attempt", resp["message"])
}

func TestAttemptHandler_GradeAttempt_InvalidAttemptID(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	gradeReq := models.AttemptUpdateDTO{}
	body, _ := json.Marshal(gradeReq)

	req := httptest.NewRequest(http.MethodPost, "/admin/attempt/grade/invalid", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempt/grade/{attemptID}", handler.GradeAttempt).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GradeAttempt", mock.Anything, mock.Anything)
}

func TestAttemptHandler_GradeAttempt_InvalidBody(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	attemptID := uint(1)
	invalidBody := []byte(`{"score": "not-a-number"}`)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/admin/attempt/grade/%d", attemptID), bytes.NewBuffer(invalidBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempt/grade/{attemptID:[0-9]+}", handler.GradeAttempt).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GradeAttempt", mock.Anything, mock.Anything)
}

func TestAttemptHandler_GradeAttempt_ServiceError(t *testing.T) {
	mockService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)
	handler := NewAttemptHandler(mockService, logger)

	attemptID := uint(1)
	gradeReq := models.AttemptUpdateDTO{Score: 90.0}
	body, _ := json.Marshal(gradeReq)

	mockService.On("GradeAttempt", gradeReq, attemptID).Return(errors.New("grading failed"))

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/admin/attempt/grade/%d", attemptID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/attempt/grade/{attemptID:[0-9]+}", handler.GradeAttempt).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}
