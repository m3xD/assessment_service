package rest

import (
	"assessment_service/internal/student/service"
	"assessment_service/internal/util"
	"encoding/base64"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type StudentHandler struct {
	studentService service.StudentService
}

func NewStudentHandler(studentService service.StudentService) *StudentHandler {
	return &StudentHandler{studentService: studentService}
}

func (h *StudentHandler) GetAvailableAssessments(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Parse pagination parameters
	params := util.GetPaginationParams(r)

	// Get available assessments
	assessments, total, err := h.studentService.GetAvailableAssessments(userID.(uint), params)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch available assessments",
		}, http.StatusInternalServerError)
		return
	}

	// Prepare pagination response
	result := util.CreatePaginationResponse(assessments, total, params)

	util.ResponseInterface(w, result, http.StatusOK)
}

func (h *StudentHandler) StartAssessment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get assessment ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	// Start assessment
	attempt, questions, settings, assessment, err := h.studentService.StartAssessment(userID.(uint), uint(id))
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to start assessment",
		}, http.StatusInternalServerError)
		return
	}

	// Calculate end time
	endsAt := attempt.StartedAt.Add(time.Duration(assessment.Duration) * time.Minute)

	// Create response
	response := map[string]interface{}{
		"attemptId":    attempt.ID,
		"assessmentId": attempt.AssessmentID,
		"title":        assessment.Title,
		"duration":     assessment.Duration,
		"timeLimit":    settings.TimeLimitEnforced,
		"endsAt":       endsAt,
		"questions":    questions,
		"settings":     settings,
	}

	util.ResponseInterface(w, response, http.StatusOK)
}

func (h *StudentHandler) GetAssessmentResultsHistory(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get assessment ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	// Get results
	results, err := h.studentService.GetAssessmentResultsHistory(userID.(uint), uint(id))
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch assessment results",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, results, http.StatusOK)
}

func (h *StudentHandler) GetAttemptDetails(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get attempt ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	// Get attempt details
	details, err := h.studentService.GetAttemptDetails(uint(id), userID.(uint))
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch attempt details",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, details, http.StatusOK)
}

func (h *StudentHandler) SaveAnswer(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get attempt ID from path
	attemptID, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	// Parse request
	var req struct {
		QuestionID uint        `json:"questionId" binding:"required"`
		Answer     interface{} `json:"answer" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	// Convert answer to string based on its type
	var answerStr string
	switch v := req.Answer.(type) {
	case string:
		answerStr = v
	case bool:
		if v {
			answerStr = "true"
		} else {
			answerStr = "false"
		}
	default:
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid answer type, must be boolean or string",
		}, http.StatusBadRequest)
		return
	}

	// Save answer
	err = h.studentService.SaveAnswer(uint(attemptID), req.QuestionID, answerStr, userID.(uint))
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to save answer",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseMap(w, map[string]interface{}{
		"status":  "SUCCESS",
		"message": "Answer saved successfully",
	}, http.StatusOK)
}

func (h *StudentHandler) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get attempt ID from path
	attemptID, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	// Submit assessment
	result, err := h.studentService.SubmitAssessment(uint(attemptID), userID.(uint))
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to submit assessment",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, result, http.StatusOK)
}

func (h *StudentHandler) SubmitMonitorEvent(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["id"]
	if !exists {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get attempt ID from path
	attemptID, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	// Parse request
	var req struct {
		EventType string                 `json:"eventType" binding:"required"`
		Timestamp string                 `json:"timestamp" binding:"required"`
		Details   map[string]interface{} `json:"details"`
		ImageData string                 `json:"imageData"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	// Decode image data if provided
	var imageData []byte
	var decodeErr error
	if req.ImageData != "" {
		imageData, decodeErr = base64.StdEncoding.DecodeString(req.ImageData)
		if decodeErr != nil {
			util.ResponseMap(w, map[string]interface{}{
				"status":  "BAD_REQUEST",
				"message": "Invalid image data",
			}, http.StatusBadRequest)
			return
		}
	}

	// Submit event
	result, err := h.studentService.SubmitMonitorEvent(uint(attemptID), req.EventType, req.Details, imageData, userID.(uint))
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to submit monitor event",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, result, http.StatusOK)
}
