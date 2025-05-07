package rest

import (
	models "assessment_service/internal/model"
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

// --- Mock QuestionService ---
type MockQuestionService struct {
	mock.Mock
}

func (m *MockQuestionService) AddQuestion(assessmentID uint, question *models.Question) (*models.Question, error) {
	args := m.Called(assessmentID, question)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Question), args.Error(1)
}
func (m *MockQuestionService) GetQuestionsByAssessment(assessmentID uint) ([]models.Question, error) {
	args := m.Called(assessmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Question), args.Error(1)
}
func (m *MockQuestionService) UpdateQuestion(questionID uint, questionData map[string]interface{}) (*models.Question, error) {
	args := m.Called(questionID, questionData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Question), args.Error(1)
}
func (m *MockQuestionService) DeleteQuestion(questionID uint) error {
	args := m.Called(questionID)
	return args.Error(0)
}

// --- Test Cases ---

func TestQuestionHandler_AddQuestion(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	assessmentID := uint(1)
	questionReq := map[string]interface{}{
		"type":          "multiple-choice",
		"text":          "Test Question?",
		"correctAnswer": "a",
		"points":        5.0, // JSON number
		"options": []struct {
			ID   int    `json:"id"`
			Text string `json:"text"`
		}{
			{1, "Option A"},
			{2, "Option B"},
		},
	}
	body, _ := json.Marshal(questionReq)

	expectedQuestion := &models.Question{
		ID:            10, // ID được gán bởi service/repo
		AssessmentID:  assessmentID,
		Type:          "multiple-choice",
		Text:          "Test Question?",
		CorrectAnswer: "a",
		Points:        5,
		Options: []models.QuestionOption{
			{OptionID: "a", Text: "Option A"},
			{OptionID: "b", Text: "Option B"},
		},
	}

	// Expect service AddQuestion to be called
	mockService.On("AddQuestion", assessmentID, mock.Anything).Return(expectedQuestion, nil)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/questions", assessmentID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/questions", handler.AddQuestion).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertExpectations(t)

	var respQuestion models.Question
	err := json.Unmarshal(rr.Body.Bytes(), &respQuestion)
	assert.NoError(t, err)
	assert.Equal(t, expectedQuestion.ID, respQuestion.ID)
	assert.Equal(t, expectedQuestion.Text, respQuestion.Text)
	assert.Len(t, respQuestion.Options, 2)
}

func TestQuestionHandler_AddQuestion_InvalidAssessmentID(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	body := []byte(`{"type":"essay", "text":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/assessments/invalid/questions", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id}/questions", handler.AddQuestion).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "AddQuestion", mock.Anything, mock.Anything)
}

func TestQuestionHandler_AddQuestion_InvalidInput(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	assessmentID := uint(1)
	invalidBody := []byte(`{"type":"essay"`) // Invalid JSON

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/questions", assessmentID), bytes.NewBuffer(invalidBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/questions", handler.AddQuestion).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "AddQuestion", mock.Anything, mock.Anything)
}

func TestQuestionHandler_AddQuestion_ServiceError(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	assessmentID := uint(1)
	questionReq := map[string]interface{}{"type": "essay", "text": "test", "points": 5.0}
	body, _ := json.Marshal(questionReq)
	serviceError := errors.New("service add error")

	mockService.On("AddQuestion", assessmentID, mock.Anything).Return(nil, serviceError)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/assessments/%d/questions", assessmentID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/questions", handler.AddQuestion).Methods(http.MethodPost)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestQuestionHandler_GetQuestionsByAssessment(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	assessmentID := uint(1)
	expectedQuestions := []models.Question{
		{ID: 1, Text: "Q1"}, {ID: 2, Text: "Q2"},
	}

	mockService.On("GetQuestionsByAssessment", assessmentID).Return(expectedQuestions, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/assessments/%d/questions", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/questions", handler.GetQuestionsByAssessment).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var respQuestions []models.Question
	err := json.Unmarshal(rr.Body.Bytes(), &respQuestions)
	assert.NoError(t, err)
	assert.Equal(t, expectedQuestions, respQuestions)
}

func TestQuestionHandler_GetQuestionsByAssessment_InvalidID(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/assessments/invalid/questions", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id}/questions", handler.GetQuestionsByAssessment).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetQuestionsByAssessment", mock.Anything)
}

func TestQuestionHandler_GetQuestionsByAssessment_ServiceError(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	assessmentID := uint(1)
	serviceError := errors.New("service get error")

	mockService.On("GetQuestionsByAssessment", assessmentID).Return(nil, serviceError)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/assessments/%d/questions", assessmentID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{id:[0-9]+}/questions", handler.GetQuestionsByAssessment).Methods(http.MethodGet)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestQuestionHandler_UpdateQuestion(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	assessmentID := uint(1)
	questionID := uint(10)
	updateReq := map[string]interface{}{
		"text":   "Updated Text",
		"points": 12.0,
	}
	body, _ := json.Marshal(updateReq)

	updatedQuestion := &models.Question{
		ID:           questionID,
		AssessmentID: assessmentID,
		Text:         "Updated Text",
		Points:       12,
	}

	mockService.On("UpdateQuestion", questionID, mock.Anything).Return(updatedQuestion, nil)

	reqPath := fmt.Sprintf("/assessments/%d/questions/%d", assessmentID, questionID)
	req := httptest.NewRequest(http.MethodPut, reqPath, bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{assessmentId:[0-9]+}/questions/{questionId:[0-9]+}", handler.UpdateQuestion).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var respQuestion models.Question
	err := json.Unmarshal(rr.Body.Bytes(), &respQuestion)
	assert.NoError(t, err)
	assert.Equal(t, updatedQuestion.ID, respQuestion.ID)
	assert.Equal(t, updatedQuestion.Text, respQuestion.Text)
	assert.Equal(t, updatedQuestion.Points, respQuestion.Points)
}

func TestQuestionHandler_UpdateQuestion_InvalidQuestionID(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	body := []byte(`{"text":"update"}`)
	req := httptest.NewRequest(http.MethodPut, "/assessments/1/questions/invalid", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{assessmentId:[0-9]+}/questions/{questionId}", handler.UpdateQuestion).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "UpdateQuestion", mock.Anything, mock.Anything)
}

func TestQuestionHandler_UpdateQuestion_InvalidInput(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	invalidBody := []byte(`{"points":"not-a-number"}`)
	req := httptest.NewRequest(http.MethodPut, "/assessments/1/questions/10", bytes.NewBuffer(invalidBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{assessmentId:[0-9]+}/questions/{questionId:[0-9]+}", handler.UpdateQuestion).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "UpdateQuestion", mock.Anything, mock.Anything)
}

func TestQuestionHandler_UpdateQuestion_ServiceError(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	questionID := uint(10)
	updateReq := map[string]interface{}{"text": "Update"}
	body, _ := json.Marshal(updateReq)
	serviceError := errors.New("service update error")

	mockService.On("UpdateQuestion", questionID, mock.Anything).Return(nil, serviceError)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/assessments/1/questions/%d", questionID), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{assessmentId:[0-9]+}/questions/{questionId:[0-9]+}", handler.UpdateQuestion).Methods(http.MethodPut)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestQuestionHandler_DeleteQuestion(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	questionID := uint(10)

	mockService.On("DeleteQuestion", questionID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/assessments/1/questions/%d", questionID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{assessmentId:[0-9]+}/questions/{questionId:[0-9]+}", handler.DeleteQuestion).Methods(http.MethodDelete)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "OK", resp["status"])
	assert.Equal(t, "Question deleted successfully", resp["message"])
}

func TestQuestionHandler_DeleteQuestion_InvalidQuestionID(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodDelete, "/assessments/1/questions/invalid", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{assessmentId:[0-9]+}/questions/{questionId}", handler.DeleteQuestion).Methods(http.MethodDelete)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "DeleteQuestion", mock.Anything)
}

func TestQuestionHandler_DeleteQuestion_ServiceError(t *testing.T) {
	mockService := new(MockQuestionService)
	logger := zaptest.NewLogger(t)
	handler := NewQuestionHandler(mockService, logger)

	questionID := uint(10)
	serviceError := errors.New("service delete error")

	mockService.On("DeleteQuestion", questionID).Return(serviceError)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/assessments/1/questions/%d", questionID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/assessments/{assessmentId:[0-9]+}/questions/{questionId:[0-9]+}", handler.DeleteQuestion).Methods(http.MethodDelete)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}
