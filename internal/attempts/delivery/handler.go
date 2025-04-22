package delivery

import (
	"assessment_service/internal/attempts/service"
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AttemptHandler struct {
	attemptService service.AttemptService
	log            *zap.Logger
}

func NewAttemptHandler(attemptService service.AttemptService, log *zap.Logger) *AttemptHandler {
	return &AttemptHandler{
		attemptService: attemptService,
		log:            log,
	}
}

func (h *AttemptHandler) GetListAttemptByUserAndAssessment(w http.ResponseWriter, r *http.Request) {
	userID, exists := r.Context().Value("user").(jwt.MapClaims)["userID"]
	if !exists {
		h.log.Error("[GetListAttemptByUserAndAssessment] userID not found in context")
		util.ResponseMap(w, map[string]interface{}{
			"status":  "UNAUTHORIZED",
			"message": "User ID not found in context",
		}, http.StatusUnauthorized)
		return
	}
	// parse userID to unit
	userIDUnit, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		h.log.Error("[GetListAttemptByUserAndAssessment] failed to convert userID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to convert userID",
		}, http.StatusInternalServerError)
		return
	}

	// Get attempt ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["assessmentID"], 10, 32)
	if err != nil {
		h.log.Error("[GetListAttemptByUserAndAssessment] invalid assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	params := util.GetPaginationParams(r)

	if params.Limit == 0 {
		params.Limit = 10
	}

	attempts, total, err := h.attemptService.GetListAttemptByUserAndAssessment(uint(userIDUnit), uint(id), params)
	if err != nil {
		h.log.Error("[GetListAttemptByUserAndAssessment] failed to get attempts", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to get attempts",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, util.CreatePaginationResponse(attempts, total, params), http.StatusOK)
}

func (h *AttemptHandler) GetAttemptDetail(w http.ResponseWriter, r *http.Request) {
	// Get attempt ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["attemptID"], 10, 32)
	if err != nil {
		h.log.Error("[GetAttemptDetail] invalid attempt ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	attempt, err := h.attemptService.GetAttemptDetail(uint(id))
	if err != nil {
		h.log.Error("[GetAttemptDetail] failed to get attempt detail", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to get attempt detail",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, attempt, http.StatusOK)
}

func (h *AttemptHandler) GradeAttempt(w http.ResponseWriter, r *http.Request) {
	// Get attempt ID from path
	id, err := strconv.ParseUint(mux.Vars(r)["attemptID"], 10, 32)
	if err != nil {
		h.log.Error("[GradeAttempt] invalid attempt ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid attempt ID",
		}, http.StatusBadRequest)
		return
	}

	var newAttempt models.Attempt
	if err := json.NewDecoder(r.Body).Decode(&newAttempt); err != nil {
		h.log.Error("[GradeAttempt] failed to parse request body", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Failed to parse request body",
		}, http.StatusBadRequest)
		return
	}

	newAttempt.ID = uint(id)

	err = h.attemptService.GradeAttempt(newAttempt)
	if err != nil {
		h.log.Error("[GradeAttempt] failed to grade attempt", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to grade attempt",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseMap(w, map[string]interface{}{
		"status":  "OK",
		"message": "Successfully graded attempt",
	}, http.StatusOK)
}
