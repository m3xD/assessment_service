package rest

import (
	// Không import mock service nữa
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time" // Cần thiết cho LogSuspiciousActivity

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require" // Dùng require khi cần
)

// --- Mock AnalyticsService (Định nghĩa trực tiếp) ---
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) GetUserActivityAnalytics() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetAssessmentPerformanceAnalytics() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) ReportActivity(activity *models.Activity) error {
	args := m.Called(activity)
	return args.Error(0)
}
func (m *MockAnalyticsService) TrackAssessmentSession(sessionData *models.SessionData) error {
	args := m.Called(sessionData)
	return args.Error(0)
}
func (m *MockAnalyticsService) LogSuspiciousActivity(activity *models.SuspiciousActivity) error {
	args := m.Called(activity)
	return args.Error(0)
}
func (m *MockAnalyticsService) GetDashboardSummary() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetActivityTimeline() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetSystemStatus() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetSuspiciousActivity(userID uint, attemptID uint, params util.PaginationParams) ([]models.SuspiciousActivity, int64, error) {
	args := m.Called(userID, attemptID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.SuspiciousActivity), args.Get(1).(int64), args.Error(2)
}

// Helper function to create a request with context containing JWT claims
func createRequestWithActivityClaims(method, url string, body []byte, claims jwt.MapClaims) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	if claims != nil {
		// Key "user" phải khớp với key được sử dụng trong middleware AuthMiddleware
		ctx := context.WithValue(req.Context(), "user", claims)
		req = req.WithContext(ctx)
	}
	return req
}

func TestAnalyticsHandler_GetUserActivityAnalytics(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	expectedAnalytics := map[string]interface{}{"daily": 100.0}
	mockService.On("GetUserActivityAnalytics").Return(expectedAnalytics, nil)

	req := httptest.NewRequest(http.MethodGet, "/analytics/user-activity", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/analytics/user-activity", handler.GetUserActivityAnalytics).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, expectedAnalytics, resp)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetUserActivityAnalytics_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	mockService.On("GetUserActivityAnalytics").Return(nil, errors.New("service error"))

	req := httptest.NewRequest(http.MethodGet, "/analytics/user-activity", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/analytics/user-activity", handler.GetUserActivityAnalytics).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetAssessmentPerformanceAnalytics(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	expectedAnalytics := map[string]interface{}{"performance": "good"}
	mockService.On("GetAssessmentPerformanceAnalytics").Return(expectedAnalytics, nil)

	req := httptest.NewRequest(http.MethodGet, "/analytics/assessment-performance", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/assessment-performance", handler.GetAssessmentPerformanceAnalytics)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, expectedAnalytics, resp)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetAssessmentPerformanceAnalytics_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	mockService.On("GetAssessmentPerformanceAnalytics").Return(nil, errors.New("perf error"))

	req := httptest.NewRequest(http.MethodGet, "/analytics/assessment-performance", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/assessment-performance", handler.GetAssessmentPerformanceAnalytics)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_ReportActivity(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	assessmentID := uint(1)
	activityReq := models.Activity{
		Action:       "VIEW_PAGE",
		AssessmentID: &assessmentID,
		Details:      "View homepage",
	}
	body, _ := json.Marshal(activityReq)

	// Sử dụng string cho userID trong claims như trong handler gốc
	claims := jwt.MapClaims{"id": "123"} // Handler gốc lấy "id" và convert

	mockService.On("ReportActivity", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Activity)
		arg.ID = 1       // Simulate ID assignment
		arg.UserID = 123 // Gán UserID để kiểm tra response
	})

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/activity", body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/activity", handler.ReportActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertExpectations(t)

	var resp models.Activity
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, uint(123), resp.UserID) // Kiểm tra UserID từ context
}

func TestAnalyticsHandler_ReportActivity_NoUserIDInContext(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	activityReq := map[string]interface{}{"action": "TEST"}
	body, _ := json.Marshal(activityReq)

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/activity", body, nil) // Không có claims
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/activity", handler.ReportActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockService.AssertNotCalled(t, "ReportActivity", mock.Anything)
}

func TestAnalyticsHandler_ReportActivity_InvalidInput(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	invalidBody := []byte(`{"action":`) // Invalid JSON
	claims := jwt.MapClaims{"id": "123"}

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/activity", invalidBody, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/activity", handler.ReportActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "ReportActivity", mock.Anything)
}

func TestAnalyticsHandler_ReportActivity_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	activityReq := map[string]interface{}{"action": "TEST"}
	body, _ := json.Marshal(activityReq)
	claims := jwt.MapClaims{"id": "123"}

	mockService.On("ReportActivity", mock.Anything).Return(errors.New("report error"))

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/activity", body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/activity", handler.ReportActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_TrackAssessmentSession(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	assessmentID := uint(1)
	sessionReq := map[string]interface{}{
		"action":    "SESSION_START",
		"userAgent": "test-agent",
	}
	body, _ := json.Marshal(sessionReq)
	claims := jwt.MapClaims{"id": "123"} // Handler gốc lấy "id"

	mockService.On("TrackAssessmentSession", mock.MatchedBy(func(sd *models.SessionData) bool {
		return sd.UserID == uint(123) && sd.AssessmentID == assessmentID && sd.Action == "SESSION_START"
	})).Return(nil)

	req := createRequestWithActivityClaims(http.MethodPost, fmt.Sprintf("/analytics/assessments/%d/session", assessmentID), body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/assessments/{id:[0-9]+}/session", handler.TrackAssessmentSession)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertExpectations(t)

	// Kiểm tra response body khớp với request body (vì handler trả về sessionData đã nhận)
	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, sessionReq["action"], resp["Action"]) // So sánh các trường quan trọng
	assert.Equal(t, sessionReq["userAgent"], resp["UserAgent"])
	assert.Equal(t, float64(123), resp["UserID"]) // GORM có thể trả về float64
	assert.Equal(t, float64(assessmentID), resp["AssessmentID"])
}

func TestAnalyticsHandler_TrackAssessmentSession_InvalidID(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)
	claims := jwt.MapClaims{"id": "123"}
	body := []byte(`{"action":"START"}`)

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/assessments/invalid/session", body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/assessments/{id}/session", handler.TrackAssessmentSession) // Route không có regex
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "TrackAssessmentSession", mock.Anything)
}

func TestAnalyticsHandler_TrackAssessmentSession_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)
	assessmentID := uint(1)
	sessionReq := map[string]interface{}{"action": "START"}
	body, _ := json.Marshal(sessionReq)
	claims := jwt.MapClaims{"id": "123"}

	mockService.On("TrackAssessmentSession", mock.Anything).Return(errors.New("track error"))

	req := createRequestWithActivityClaims(http.MethodPost, fmt.Sprintf("/analytics/assessments/%d/session", assessmentID), body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/assessments/{id:[0-9]+}/session", handler.TrackAssessmentSession)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_LogSuspiciousActivity(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	suspiciousReq := map[string]interface{}{
		"attemptID":    "100",
		"assessmentId": "1",
		"type":         "TAB_SWITCH",
		"details":      "Switched too many times",
		"timestamp":    time.Now().Format(time.RFC3339), // Thêm timestamp hợp lệ
	}
	body, _ := json.Marshal(suspiciousReq)
	claims := jwt.MapClaims{"userID": "123"} // Handler lấy userID từ claims

	mockService.On("LogSuspiciousActivity", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.SuspiciousActivity)
		arg.ID = 99 // Simulate ID assignment
	})

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/suspicious", body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/suspicious", handler.LogSuspiciousActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertExpectations(t)

	var resp models.SuspiciousActivity
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, uint(99), resp.ID)
	assert.Equal(t, "LOW", resp.Severity) // Kiểm tra severity đã được set
}

func TestAnalyticsHandler_LogSuspiciousActivity_InvalidInput(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	invalidBody := []byte(`{"type":`) // JSON không hợp lệ
	claims := jwt.MapClaims{"userID": "123"}

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/suspicious", invalidBody, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/suspicious", handler.LogSuspiciousActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "LogSuspiciousActivity", mock.Anything)
}

func TestAnalyticsHandler_LogSuspiciousActivity_InvalidTimestamp(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	suspiciousReq := map[string]interface{}{
		"attemptID": "100", "assessmentId": "1", "type": "TEST", "timestamp": "invalid-date",
	}
	body, _ := json.Marshal(suspiciousReq)
	claims := jwt.MapClaims{"userID": "123"}

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/suspicious", body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/suspicious", handler.LogSuspiciousActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code) // Expect Bad Request due to timestamp format
	mockService.AssertNotCalled(t, "LogSuspiciousActivity", mock.Anything)
}

func TestAnalyticsHandler_LogSuspiciousActivity_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	suspiciousReq := map[string]interface{}{"attemptID": "100", "assessmentId": "1", "type": "TEST"}
	body, _ := json.Marshal(suspiciousReq)
	claims := jwt.MapClaims{"userID": "123"}

	mockService.On("LogSuspiciousActivity", mock.Anything).Return(errors.New("log error"))

	req := createRequestWithActivityClaims(http.MethodPost, "/analytics/suspicious", body, claims)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/analytics/suspicious", handler.LogSuspiciousActivity)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetDashboardSummary(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)
	expectedSummary := map[string]interface{}{"totalUsers": 1000.0}

	mockService.On("GetDashboardSummary").Return(expectedSummary, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard/summary", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/dashboard/summary", handler.GetDashboardSummary)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, expectedSummary, resp)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetDashboardSummary_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	mockService.On("GetDashboardSummary").Return(nil, errors.New("summary error"))

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard/summary", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/dashboard/summary", handler.GetDashboardSummary)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetActivityTimeline(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)
	expectedTimeline := map[string]interface{}{"timeline": []string{"event1"}}

	mockService.On("GetActivityTimeline").Return(expectedTimeline, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard/activity", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/dashboard/activity", handler.GetActivityTimeline)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetActivityTimeline_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	mockService.On("GetActivityTimeline").Return(nil, errors.New("timeline error"))

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard/activity", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/dashboard/activity", handler.GetActivityTimeline)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSystemStatus(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)
	expectedStatus := map[string]interface{}{"status": "healthy"}

	mockService.On("GetSystemStatus").Return(expectedStatus, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/system/status", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/system/status", handler.GetSystemStatus)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, expectedStatus, resp)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSystemStatus_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	mockService.On("GetSystemStatus").Return(nil, errors.New("status error"))

	req := httptest.NewRequest(http.MethodGet, "/admin/system/status", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/system/status", handler.GetSystemStatus)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSuspiciousActivity(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	userID := uint(1)
	attemptID := uint(10)
	expectedActivities := []models.SuspiciousActivity{{ID: 1, UserID: userID, AttemptID: attemptID}}
	expectedTotal := int64(1)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}}

	mockService.On("GetSuspiciousActivity", userID, attemptID, expectedParams).Return(expectedActivities, expectedTotal, nil)

	reqPath := fmt.Sprintf("/admin/activity/%d/%d", userID, attemptID)
	req := httptest.NewRequest(http.MethodGet, reqPath, nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/activity/{userID:[0-9]+}/{attemptID:[0-9]+}", handler.GetSuspiciousActivity).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(expectedTotal), result["totalElements"])
	content, ok := result["content"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, content, len(expectedActivities))
}

func TestAnalyticsHandler_GetSuspiciousActivity_InvalidUserID(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/admin/activity/invalid/10", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/activity/{userID}/{attemptID:[0-9]+}", handler.GetSuspiciousActivity).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetSuspiciousActivity", mock.Anything, mock.Anything, mock.Anything)
}

func TestAnalyticsHandler_GetSuspiciousActivity_InvalidAttemptID(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/admin/activity/1/invalid", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/admin/activity/{userID:[0-9]+}/{attemptID}", handler.GetSuspiciousActivity).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetSuspiciousActivity", mock.Anything, mock.Anything, mock.Anything)
}

func TestAnalyticsHandler_GetSuspiciousActivity_ServiceError(t *testing.T) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)

	userID := uint(1)
	attemptID := uint(10)
	expectedParams := util.PaginationParams{Page: 0, Limit: 10, Offset: 0, SortBy: "created_at", SortDir: "DESC", Filters: map[string]interface{}{}}

	mockService.On("GetSuspiciousActivity", userID, attemptID, expectedParams).Return(nil, int64(0), errors.New("fetch error"))

	reqPath := fmt.Sprintf("/admin/activity/%d/%d", userID, attemptID)
	req := httptest.NewRequest(http.MethodGet, reqPath, nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/admin/activity/{userID:[0-9]+}/{attemptID:[0-9]+}", handler.GetSuspiciousActivity).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}
