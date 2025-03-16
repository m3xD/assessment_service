package rest

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/questions/service"
	"assessment_service/internal/util"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type QuestionHandler struct {
	questionService service.QuestionService
	log             *zap.Logger
}

func NewQuestionHandler(questionService service.QuestionService, log *zap.Logger) *QuestionHandler {
	return &QuestionHandler{questionService: questionService, log: log}
}

func (h *QuestionHandler) AddQuestion(w http.ResponseWriter, r *http.Request) {
	assessmentID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("Invalid assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Type          string                  `json:"type" binding:"required,oneof=multiple-choice true-false essay"`
		Text          string                  `json:"text" binding:"required"`
		Options       []models.QuestionOption `json:"options"`
		CorrectAnswer string                  `json:"correctAnswer"`
		Points        float64                 `json:"points" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("Invalid input", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	question := &models.Question{
		Type:          req.Type,
		Text:          req.Text,
		Options:       req.Options,
		CorrectAnswer: req.CorrectAnswer,
		Points:        req.Points,
	}

	question, err = h.questionService.AddQuestion(uint(assessmentID), question)
	if err != nil {
		h.log.Error("Failed to add question", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to add question",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, question, http.StatusCreated)
}

func (h *QuestionHandler) GetQuestionsByAssessment(w http.ResponseWriter, r *http.Request) {
	assessmentID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		h.log.Error("Invalid assessment ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid assessment ID",
		}, http.StatusBadRequest)
		return
	}

	questions, err := h.questionService.GetQuestionsByAssessment(uint(assessmentID))
	if err != nil {
		h.log.Error("Failed to get questions", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to get questions",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, questions, http.StatusOK)
}

func (h *QuestionHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := strconv.ParseUint(mux.Vars(r)["questionId"], 10, 32)
	if err != nil {
		h.log.Error("Invalid question ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid question ID",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		Text          string                  `json:"text"`
		Options       []models.QuestionOption `json:"options"`
		CorrectAnswer interface{}             `json:"correctAnswer"`
		Points        float64                 `json:"points"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("Invalid input", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid input",
		}, http.StatusBadRequest)
		return
	}

	// Convert request to map for partial updates
	questionData := make(map[string]interface{})

	if req.Text != "" {
		questionData["text"] = req.Text
	}

	if req.Points != 0 {
		questionData["points"] = req.Points
	}

	if req.CorrectAnswer != nil {
		questionData["correctAnswer"] = req.CorrectAnswer
	}

	if req.Options != nil {
		questionData["options"] = req.Options
	}

	question, err := h.questionService.UpdateQuestion(uint(questionID), questionData)
	if err != nil {
		h.log.Error("Failed to update question", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to update question",
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseInterface(w, question, http.StatusOK)
}

func (h *QuestionHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := strconv.ParseUint(mux.Vars(r)["questionId"], 10, 32)
	if err != nil {
		h.log.Error("Invalid question ID", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "BAD_REQUEST",
			"message": "Invalid question ID",
		}, http.StatusBadRequest)
		return
	}

	err = h.questionService.DeleteQuestion(uint(questionID))
	if err != nil {
		h.log.Error("Failed to delete question", zap.Error(err))
		util.ResponseMap(w, map[string]interface{}{
			"status":  "ERROR",
			"message": "Failed to delete question",
			"path":    r.URL.Path,
		}, http.StatusInternalServerError)
		return
	}

	util.ResponseMap(w, map[string]interface{}{
		"status":  "OK",
		"message": "Question deleted successfully",
	}, http.StatusOK)
}
