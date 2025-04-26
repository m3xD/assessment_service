package rest

import (
	"assessment_service/internal/assessments/service"
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type AssessmentHandler struct {
	assessmentService service.AssessmentService
	log               *zap.Logger
}

func NewAssessmentHandler(assessmentService service.AssessmentService, log *zap.Logger) *AssessmentHandler {
	return &AssessmentHandler{assessmentService: assessmentService, log: log}
}

func (h *AssessmentHandler) CreateAssessment(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !ok {
		h.log.Error("[CreateAssessment] Failed to get user ID from token")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "Unauthorized",
		}, http.StatusUnauthorized)
		return
	}

	// convert string to id
	id, err := strconv.ParseUint(claims.(string), 10, 32)
	if err != nil {
		h.log.Error("[CreateAssessment] Failed to parse user ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "Unauthorized",
		}, http.StatusUnauthorized)
		return
	}

	var req struct {
		Title        string  `json:"title" binding:"required"`
		Subject      string  `json:"subject" binding:"required"`
		Description  string  `json:"description"`
		Duration     int     `json:"duration" binding:"required"`
		DueDate      string  `json:"dueDate"`
		Id           uint    `json:"id"`
		PassingScore float64 `json:"passingScore" binding:"required"`
		Status       string  `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("[CreateAssessment] Failed to decode request body", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	assessment := &models.Assessment{
		Title:        req.Title,
		Subject:      req.Subject,
		Description:  req.Description,
		Duration:     req.Duration,
		CreatedByID:  uint(id),
		PassingScore: req.PassingScore,
		Status:       req.Status,
	}

	if req.DueDate != "" {
		dueDate, err := time.Parse(time.RFC3339, req.DueDate)
		if err == nil {
			assessment.DueDate = &dueDate
		}
	}

	err = h.assessmentService.Create(assessment)
	if err != nil {
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to create assessment",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, assessment, http.StatusCreated)
}

func (h *AssessmentHandler) GetAssessmentById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[GetAssessmentById] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	assessment, err := h.assessmentService.GetByID(uint(id))
	if err != nil {
		h.log.Error("[GetAssessmentById] Failed to fetch assessment", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "NOT_FOUND",
			"message": "Assessment not found",
		}, http.StatusNotFound)
		return
	}

	util.ResponseInterface(w, assessment, http.StatusOK)
}

func (h *AssessmentHandler) UpdateAssessment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[UpdateAssessment] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Title        string  `json:"title"`
		Subject      string  `json:"subject"`
		Description  string  `json:"description"`
		Duration     float64 `json:"duration"`
		DueDate      string  `json:"dueDate"`
		PassingScore float64 `json:"passingScore"`
		Status       string  `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("[UpdateAssessment] Failed to decode request body", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	// Convert request to map for partial updates
	assessmentData := make(map[string]interface{})

	if req.Title != "" {
		assessmentData["title"] = req.Title
	}

	if req.Subject != "" {
		assessmentData["subject"] = req.Subject
	}

	if req.Description != "" {
		assessmentData["description"] = req.Description
	}

	if req.Duration != 0 {
		assessmentData["duration"] = req.Duration
	}

	if req.DueDate != "" {
		assessmentData["dueDate"] = req.DueDate
	}

	if req.PassingScore != 0 {
		assessmentData["passingScore"] = req.PassingScore
	}

	if req.Status != "" {
		assessmentData["status"] = req.Status
	}

	assessment, err := h.assessmentService.Update(uint(id), assessmentData)
	if err != nil {
		h.log.Error("[UpdateAssessment] Failed to update assessment", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to update assessment",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, assessment, http.StatusOK)
}

func (h *AssessmentHandler) DeleteAssessment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[DeleteAssessment] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	err = h.assessmentService.Delete(uint(id))
	if err != nil {
		h.log.Error("[DeleteAssessment] Failed to delete assessment", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to delete assessment",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseMap(w, map[string]interface{}{
		"status":  "OK",
		"message": "Assessment deleted successfully",
	}, http.StatusOK)
}

func (h *AssessmentHandler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	params := util.GetPaginationParams(r)

	// Add filters
	if subject := r.URL.Query().Get("subject"); subject != "" {
		params.Filters["subject"] = subject
	}

	if status := r.URL.Query().Get("status"); status != "" {
		params.Filters["status"] = status
	}

	// Get assessments with pagination
	assessments, total, err := h.assessmentService.List(params)
	if err != nil {
		h.log.Error("[ListAssessments] Failed to fetch assessments", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch assessments",
		}, http.StatusInternalServerError)
		return
	}

	// Prepare pagination response
	result := util.CreatePaginationResponse(assessments, total, params)

	util.ResponseInterface(w, result, http.StatusOK)
}

func (h *AssessmentHandler) GetRecentAssessments(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 5 // Default limit
	}

	assessments, err := h.assessmentService.GetRecentAssessments(limit)
	if err != nil {
		h.log.Error("[GetRecentAssessments] Failed to fetch recent assessments", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch recent assessments",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, assessments, http.StatusOK)
}

func (h *AssessmentHandler) GetAssessmentStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.assessmentService.GetStatistics()
	if err != nil {
		h.log.Error("[GetAssessmentStatistics] Failed to fetch assessment statistics", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch assessment statistics",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, stats, http.StatusOK)
}

func (h *AssessmentHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[UpdateSettings] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	var req models.AssessmentSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("[UpdateSettings] Failed to decode request body", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	err = h.assessmentService.UpdateSettings(uint(id), &req)
	if err != nil {
		h.log.Error("[UpdateSettings] Failed to update assessment settings", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to update assessment settings",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseMap(w, map[string]interface{}{
		"id":       id,
		"settings": req,
	}, http.StatusOK)
}

func (h *AssessmentHandler) GetAssessmentResults(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[GetAssessmentResults] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	params := util.GetPaginationParams(r)

	// Add filters
	if user := r.URL.Query().Get("user"); user != "" {
		params.Filters["user"] = user
	}

	// Get assessment results with pagination
	results, total, err := h.assessmentService.GetResults(uint(id), params)
	if err != nil {
		h.log.Error("[GetAssessmentResults] Failed to fetch assessment results", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch assessment results",
		}, http.StatusInternalServerError)
		return
	}

	// Prepare pagination response
	result := util.CreatePaginationResponse(results, total, params)

	util.ResponseInterface(w, result, http.StatusOK)
}

func (h *AssessmentHandler) PublishAssessment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[PublishAssessment] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	assessment, err := h.assessmentService.Publish(uint(id))
	if err != nil {
		h.log.Error("[PublishAssessment] Failed to publish assessment", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to publish assessment",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, assessment, http.StatusOK)
}

func (h *AssessmentHandler) DuplicateAssessment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("[DuplicateAssessment] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		NewTitle      string `json:"newTitle"`
		CopyQuestions bool   `json:"copyQuestions"`
		CopySettings  bool   `json:"copySettings"`
		SetAsDraft    bool   `json:"setAsDraft"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("[DuplicateAssessment] Failed to decode request body", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	assessment, err := h.assessmentService.Duplicate(uint(id), req.NewTitle, req.CopyQuestions, req.CopySettings, req.SetAsDraft)
	if err != nil {
		h.log.Error("[DuplicateAssessment] Failed to duplicate assessment", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to duplicate assessment",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, assessment, http.StatusCreated)
}

func (h *AssessmentHandler) GetAssessmentWithUserHasAttempt(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["assessmentID"], 10, 32)
	if err != nil {
		h.log.Error("[GetAssessmentWithUserHasAttempt] Failed to parse assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	params := util.GetPaginationParams(r)

	assessment, user, total, err := h.assessmentService.GetAssessmentDetailWithUser(uint(id), params)
	if err != nil {
		h.log.Error("[GetAssessmentWithUserHasAttempt] Failed to fetch assessment", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to fetch assessment",
		}, http.StatusInternalServerError)
		return
	}

	users := util.CreatePaginationResponse(user, total, params)

	util.ResponseMap(w, map[string]interface{}{
		"assessment": assessment,
		"users":      users,
	}, http.StatusOK)
}
