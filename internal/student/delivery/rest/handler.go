package rest

import (
	"assessment_service/internal/student/service"
	"assessment_service/internal/util"
	"encoding/base64"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type StudentHandler struct {
	studentService service.StudentService
	log            *zap.Logger
}

func NewStudentHandler(studentService service.StudentService, log *zap.Logger) *StudentHandler {
	return &StudentHandler{studentService: studentService, log: log}
}

func (h *StudentHandler) GetAvailableAssessments(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[GetAvailableAssessments] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Parse pagination parameters
	params := util.GetPaginationParams(r)

	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[GetAvailableAssessments] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get available assessments
	assessments, total, err := h.studentService.GetAvailableAssessments(uint(userIDUnit), params)
	if err != nil {
		h.log.Error("[GetAvailableAssessments] failed to fetch available assessments", zap.Error(err))
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
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[StartAssessment] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[StartAssessment] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get assessment ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[StartAssessment] invalid assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	// Start assessment
	attempt, questions, settings, assessment, err := h.studentService.StartAssessment(uint(userIDUnit), uint(id))
	if err != nil {
		h.log.Error("[StartAssessment] failed to start assessment", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to start assessment " + err.Error(),
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
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[GetAssessmentResultsHistory] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[GetAssessmentResultsHistory] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get assessment ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[GetAssessmentResultsHistory] invalid assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	// Get results
	results, err := h.studentService.GetAssessmentResultsHistory(uint(userIDUnit), uint(id))
	if err != nil {
		h.log.Error("[GetAssessmentResultsHistory] failed to fetch assessment results", zap.Error(err))
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
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[GetAttemptDetails] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}
	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[GetAttemptDetails] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get attempt ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		h.log.Error("[GetAttemptDetails] invalid attempt ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	// Get attempt details
	details, err := h.studentService.GetAttemptDetails(uint(id), uint(userIDUnit))
	if err != nil {
		h.log.Error("[GetAttemptDetails] failed to fetch attempt details", zap.Error(err))
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
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[SaveAnswer] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}
	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[SaveAnswer] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get attempt ID from path
	attemptID, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		h.log.Error("[SaveAnswer] invalid attempt ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	// Parse request
	var req struct {
		QuestionID string      `json:"questionId" binding:"required"`
		Answer     interface{} `json:"answer" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("[SaveAnswer] invalid input", zap.Error(err))
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
		h.log.Error("[SaveAnswer] invalid answer type", zap.Any("answer", req.Answer))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid answer type, must be boolean or string",
		}, http.StatusBadRequest)
		return
	}

	// convert string to unit
	questionIDUnit, err := strconv.ParseUint(req.QuestionID, 10, 32)
	if err != nil {
		h.log.Error("[SaveAnswer] failed to convert questionID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert questionID",
		}, http.StatusBadRequest)
	}

	// Save answer
	err = h.studentService.SaveAnswer(uint(attemptID), uint(questionIDUnit), answerStr, uint(userIDUnit))
	if err != nil {
		h.log.Error("[SaveAnswer] failed to save answer", zap.Error(err))
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
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[SubmitAssessment] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}
	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[SubmitAssessment] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get attempt ID from path
	attemptID, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		h.log.Error("[SubmitAssessment] invalid attempt ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	// Submit assessment
	result, err := h.studentService.SubmitAssessment(uint(attemptID), uint(userIDUnit))
	if err != nil {
		h.log.Error("[SubmitAssessment] failed to submit assessment", zap.Error(err))
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
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[SubmitMonitorEvent] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}
	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[SubmitMonitorEvent] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get attempt ID from path
	attemptID, err := strconv.ParseUint(mux.Vars(r)["attemptId"], 10, 32)
	if err != nil {
		h.log.Error("[SubmitMonitorEvent] invalid attempt ID", zap.Error(err))
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
		h.log.Error("[SubmitMonitorEvent] invalid input", zap.Error(err))
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
			h.log.Error("[SubmitMonitorEvent] failed to decode image data", zap.Error(decodeErr))
			util.ResponseMap(w, map[string]interface{}{
				"status":  "BAD_REQUEST",
				"message": "Invalid image data",
			}, http.StatusBadRequest)
			return
		}
	}

	// Submit event
	result, err := h.studentService.SubmitMonitorEvent(uint(attemptID), req.EventType, req.Details, imageData, uint(userIDUnit))
	if err != nil {
		h.log.Error("[SubmitMonitorEvent] failed to submit monitor event", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to submit monitor event",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, result, http.StatusOK)
}

func (h *StudentHandler) GetAllAttemptForUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["userID"], 10, 32)
	if err != nil {
		h.log.Error("[GetListAttemptByUserAndAssessment] invalid user ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid userID",
		}, http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	params := util.GetPaginationParams(r)

	// Get all attempts for user
	attempts, total, err := h.studentService.GetAllAttemptByUserID(uint(userID), params)
	if err != nil {
		h.log.Error("[GetAllAttemptForUser] failed to fetch attempts", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch attempts",
		}, http.StatusInternalServerError)
		return
	}

	// Prepare pagination response
	result := util.CreatePaginationResponse(attempts, total, params)

	util.ResponseInterface(w, result, http.StatusOK)
}
