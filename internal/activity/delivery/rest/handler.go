package rest

import (
	"assessment_service/internal/activity/service"
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) GetUserActivityAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics, err := h.analyticsService.GetUserActivityAnalytics()
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch user activity analytics",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, analytics, http.StatusOK)
}

func (h *AnalyticsHandler) GetAssessmentPerformanceAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics, err := h.analyticsService.GetAssessmentPerformanceAnalytics()
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch assessment performance analytics",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, analytics, http.StatusOK)
}

func (h *AnalyticsHandler) ReportActivity(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	claims, _ := r.Context().Value("user").(jwt.MapClaims)
	userID, exists := claims["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	var req struct {
		Action       string     `json:"action" binding:"required"`
		AssessmentID *uint      `json:"assessmentId"`
		Details      string     `json:"details"`
		Timestamp    *time.Time `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	// Create activity
	activity := &models.Activity{
		UserID:       userID.(uint),
		Action:       req.Action,
		AssessmentID: req.AssessmentID,
		Details:      req.Details,
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	if req.Timestamp != nil {
		activity.Timestamp = *req.Timestamp
	} else {
		activity.Timestamp = time.Now()
	}

	newErr := h.analyticsService.ReportActivity(activity)
	if newErr != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to report activity",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, activity, http.StatusCreated)
}

func (h *AnalyticsHandler) TrackAssessmentSession(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	claims, _ := r.Context().Value("user").(jwt.MapClaims)
	userID, exists := claims["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get assessment ID from path
	id, newErr := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if newErr != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Action     string     `json:"action" binding:"required"`
		Timestamp  *time.Time `json:"timestamp"`
		UserAgent  string     `json:"userAgent"`
		QuestionID *uint      `json:"questionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	// Create session data
	sessionData := &models.SessionData{
		UserID:       userID.(uint),
		AssessmentID: uint(id),
		Action:       req.Action,
		UserAgent:    req.UserAgent,
	}

	if req.QuestionID != nil {
		sessionData.Details = fmt.Sprintf("Question ID: %d", *req.QuestionID)
	}

	if req.Timestamp != nil {
		sessionData.Timestamp = *req.Timestamp
	} else {
		sessionData.Timestamp = time.Now()
	}

	newErr = h.analyticsService.TrackAssessmentSession(sessionData)
	if newErr != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to track assessment session",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, sessionData, http.StatusCreated)
}

func (h *AnalyticsHandler) LogSuspiciousActivity(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	claims, _ := r.Context().Value("user").(jwt.MapClaims)
	userID, exists := claims["userID"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	var req struct {
		AssessmentID uint       `json:"assessmentId" binding:"required"`
		Type         string     `json:"type" binding:"required"`
		Details      string     `json:"details"`
		Timestamp    *time.Time `json:"timestamp"`
		UserAgent    string     `json:"userAgent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	// Create suspicious activity
	activity := &models.SuspiciousActivity{
		UserID:       userID.(uint),
		AssessmentID: req.AssessmentID,
		Type:         req.Type,
		Details:      req.Details,
	}

	if req.Timestamp != nil {
		activity.Timestamp = *req.Timestamp
	} else {
		activity.Timestamp = time.Now()
	}

	// Determine severity based on event type
	switch req.Type {
	case "TAB_SWITCHING", "MULTIPLE_FACES":
		activity.Severity = "HIGH"
	case "FACE_NOT_DETECTED", "LOOKING_AWAY":
		activity.Severity = "MEDIUM"
	default:
		activity.Severity = "LOW"
	}

	newErr := h.analyticsService.LogSuspiciousActivity(activity)
	if newErr != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to log suspicious activity",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, activity, http.StatusCreated)
}

func (h *AnalyticsHandler) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.analyticsService.GetDashboardSummary()
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch dashboard summary",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, summary, http.StatusOK)
}

func (h *AnalyticsHandler) GetActivityTimeline(w http.ResponseWriter, r *http.Request) {
	timeline, err := h.analyticsService.GetActivityTimeline()
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch activity timeline",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, timeline, http.StatusOK)
}

func (h *AnalyticsHandler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.analyticsService.GetSystemStatus()
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch system status",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, status, http.StatusOK)
}
