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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// --- Mock StudentService ---
type MockStudentService struct {
	mock.Mock
}

func (m *MockStudentService) GetAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(userID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]map[string]interface{}), args.Get(1).(int64), args.Error(2)
}
func (m *MockStudentService) StartAssessment(userID, assessmentID uint) (*models.Attempt, []models.Question, *models.AssessmentSettings, *models.Assessment, error) {
	args := m.Called(userID, assessmentID)
	// Handle nil returns carefully
	var attempt *models.Attempt
	if args.Get(0) != nil {
		attempt = args.Get(0).(*models.Attempt)
	}
	var questions []models.Question
	if args.Get(1) != nil {
		questions = args.Get(1).([]models.Question)
	}
	var settings *models.AssessmentSettings
	if args.Get(2) != nil {
		settings = args.Get(2).(*models.AssessmentSettings)
	}
	var assessment *models.Assessment
	if args.Get(3) != nil {
		assessment = args.Get(3).(*models.Assessment)
	}
	return attempt, questions, settings, assessment, args.Error(4)
}
func (m *MockStudentService) GetAssessmentResultsHistory(userID, assessmentID uint) ([]map[string]interface{}, error) {
	args := m.Called(userID, assessmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockStudentService) GetAttemptDetails(attemptID, userID uint) (*map[string]interface{}, error) {
	args := m.Called(attemptID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Return as pointer to map
	resMap := args.Get(0).(*map[string]interface{})
	return resMap, args.Error(1)
}
func (m *MockStudentService) SaveAnswer(attemptID, questionID uint, answer string, userID uint) error {
	args := m.Called(attemptID, questionID, answer, userID)
	return args.Error(0)
}
func (m *MockStudentService) SubmitAssessment(attemptID, userID uint) (*map[string]interface{}, error) {
	args := m.Called(attemptID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resMap, _ := args.Get(0).(*map[string]interface{})
	return resMap, args.Error(1)
}
func (m *MockStudentService) SubmitMonitorEvent(attemptID uint, eventType string, details map[string]interface{}, imageData []byte, userID uint) (*map[string]interface{}, error) {
	// Note: Comparing []byte might be tricky with mock.Anything, consider specific matching if needed
	args := m.Called(attemptID, eventType, details, imageData, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resMap, _ := args.Get(0).(*map[string]interface{})
	return resMap, args.Error(1)
}
func (m *MockStudentService) AutoSubmitAssessment() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockStudentService) GetAllAttemptByUserID(userID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Attempt), args.Get(1).(int64), args.Error(2)
}

// Helper function to create a request with context containing JWT claims
func createRequestWithStudentClaims(method, url string, body []byte, claims jwt.MapClaims) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	if claims != nil {
		ctx := context.WithValue(req.Context(), "user", claims) // Key "user" phải khớp
		req = req.WithContext(ctx)
	}
	return req
}

// --- Test Cases ---

func TestStudentHandler_GetAvailableAssessments(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	userID := uint(123)
	expectedAssessments := []map[string]interface{}{{"id": "1", "title": "Quiz 1"}}
	expectedTotal := int64(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}} // Default params

	claims := jwt.MapClaims{"userID": "123"} // User ID từ token
	mockService.On("GetAvailableAssessments", userID, expectedParams).Return(expectedAssessments, expectedTotal, nil)

	req := createRequestWithStudentClaims(http.MethodGet, "/student/assessments/available", nil, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/student/assessments/available", handler.GetAvailableAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(expectedTotal), result["totalElements"])
	content, ok := result["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, content, len(expectedAssessments))
}

func TestStudentHandler_GetAvailableAssessments_NoUserID(t *testing.T) {
	//mockService := new(MockStudentService)
	//logger := zaptest.NewLogger(t)
	//handler := NewStudentHandler(mockService, logger)
	//
	//req := createRequestWithStudentClaims(http.MethodGet, "/student/assessments/available", nil, nil) // No claims
	//rr := httptest.NewRecorder()
	//router := mux.NewRouter()
	//router.HandleFunc("/student/assessments/available", handler.GetAvailableAssessments).Methods(http.MethodGet)
	//router.ServeHTTP(rr, req)
	//
	//assert.Equal(t, http.StatusUnauthorized, rr.Code)
	//mockService.AssertNotCalled(t, "GetAvailableAssessments", mock.Anything, mock.Anything)
}

func TestStudentHandler_GetAvailableAssessments_ServiceError(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	userID := uint(123)
	params := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}}
	claims := jwt.MapClaims{"userID": "123"}

	mockService.On("GetAvailableAssessments", userID, params).Return(nil, int64(0), errors.New("service error"))

	req := createRequestWithStudentClaims(http.MethodGet, "/student/assessments/available", nil, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/student/assessments/available", handler.GetAvailableAssessments).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestStudentHandler_StartAssessment(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	userID := uint(123)
	assessmentID := uint(10)
	now := time.Now()
	expectedAttempt := &models.Attempt{ID: 99, UserID: userID, AssessmentID: assessmentID, StartedAt: now}
	expectedQuestions := []models.Question{{ID: 1, Text: "Q1"}}
	expectedSettings := &models.AssessmentSettings{TimeLimitEnforced: true}
	expectedAssessment := &models.Assessment{ID: assessmentID, Title: "Test Quiz", Duration: 60}

	claims := jwt.MapClaims{"userID": "123"}
	mockService.On("StartAssessment", userID, assessmentID).Return(expectedAttempt, expectedQuestions, expectedSettings, expectedAssessment, nil)

	req := createRequestWithStudentClaims(http.MethodPost, fmt.Sprintf("/student/assessments/%d/start", assessmentID), nil, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/student/assessments/{id:[0-9]+}/start", handler.StartAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(expectedAttempt.ID), resp["attemptId"])
	assert.Equal(t, float64(assessmentID), resp["assessmentId"])
	assert.Equal(t, expectedAssessment.Title, resp["title"])
	questionsResp, ok := resp["questions"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, questionsResp, 1)
}

// Thêm test case lỗi cho StartAssessment (invalid ID, service error, user not found in context)

func TestStudentHandler_GetAssessmentResultsHistory(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	userID := uint(123)
	assessmentID := uint(10)
	expectedResults := []map[string]interface{}{{"attemptId": float64(1), "score": float64(90.0)}}
	claims := jwt.MapClaims{"userID": "123"}

	mockService.On("GetAssessmentResultsHistory", userID, assessmentID).Return(expectedResults, nil)

	req := createRequestWithStudentClaims(http.MethodGet, fmt.Sprintf("/student/assessments/%d/results", assessmentID), nil, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/student/assessments/{id:[0-9]+}/results", handler.GetAssessmentResultsHistory).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedResults, resp)
}

// Thêm test case lỗi cho GetAssessmentResultsHistory

func TestStudentHandler_GetAttemptDetails(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	attemptID := uint(5)
	userID := uint(123)
	expectedDetails := map[string]interface{}{"attemptId": float64(attemptID), "status": "In Progress"}
	claims := jwt.MapClaims{"userID": "123"}

	mockService.On("GetAttemptDetails", attemptID, userID).Return(&expectedDetails, nil)

	req := createRequestWithStudentClaims(http.MethodGet, fmt.Sprintf("/student/attempts/%d", attemptID), nil, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/student/attempts/{attemptId:[0-9]+}", handler.GetAttemptDetails).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedDetails, resp)
}

// Thêm test case lỗi cho GetAttemptDetails

func TestStudentHandler_SaveAnswer(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	attemptID := uint(1)
	questionID := uint(101)
	userID := uint(123)
	answerReq := map[string]interface{}{"questionId": strconv.Itoa(int(questionID)), "answer": "true"}
	body, _ := json.Marshal(answerReq)
	claims := jwt.MapClaims{"userID": "123"}

	mockService.On("SaveAnswer", attemptID, questionID, "true", userID).Return(nil)

	req := createRequestWithStudentClaims(http.MethodPost, fmt.Sprintf("/student/attempts/%d/answers", attemptID), body, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/student/attempts/{attemptId:[0-9]+}/answers", handler.SaveAnswer).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", resp["status"])
}

// Thêm test case lỗi cho SaveAnswer

func TestStudentHandler_SubmitAssessment(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	attemptID := uint(1)
	userID := uint(123)
	expectedResult := map[string]interface{}{"completed": true, "score": 80.0}
	claims := jwt.MapClaims{"userID": "123"}

	mockService.On("SubmitAssessment", attemptID, userID).Return(&expectedResult, nil)

	req := createRequestWithStudentClaims(http.MethodPost, fmt.Sprintf("/student/attempts/%d/submit", attemptID), nil, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/student/attempts/{attemptId:[0-9]+}/submit", handler.SubmitAssessment).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, resp)
}

// Thêm test case lỗi cho SubmitAssessment

func TestStudentHandler_SubmitMonitorEvent(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	attemptID := uint(1)
	userID := uint(123)
	eventReq := map[string]interface{}{
		"eventType": "TAB_SWITCH",
		"timestamp": time.Now().Format(time.RFC3339),
		"details":   map[string]interface{}{"count": 1.0},
	}
	body, _ := json.Marshal(eventReq)
	claims := jwt.MapClaims{"userID": "123"}
	expectedResponse := map[string]interface{}{"received": true, "severity": "CRITICAL"}

	// Sử dụng mock.AnythingOfType cho imageData ([]byte) vì so sánh byte slice phức tạp
	mockService.On("SubmitMonitorEvent", attemptID, "TAB_SWITCH", mock.AnythingOfType("map[string]interface {}"), mock.AnythingOfType("[]uint8"), userID).Return(&expectedResponse, nil)

	req := createRequestWithStudentClaims(http.MethodPost, fmt.Sprintf("/student/attempts/%d/monitor", attemptID), body, claims)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/student/attempts/{attemptId:[0-9]+}/monitor", handler.SubmitMonitorEvent).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse["received"], resp["received"])
	assert.Equal(t, expectedResponse["severity"], resp["severity"])
}

// Thêm test case lỗi cho SubmitMonitorEvent

func TestStudentHandler_GetAllAttemptForUser(t *testing.T) {
	mockService := new(MockStudentService)
	logger := zaptest.NewLogger(t)
	handler := NewStudentHandler(mockService, logger)

	userID := uint(1)
	expectedAttempts := []models.Attempt{{ID: 1, UserID: userID}}
	expectedTotal := int64(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}}

	mockService.On("GetAllAttemptByUserID", userID, expectedParams).Return(expectedAttempts, expectedTotal, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/users/%d/attempts", userID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/users/{userID:[0-9]+}/attempts", handler.GetAllAttemptForUser).Methods(http.MethodGet)
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
