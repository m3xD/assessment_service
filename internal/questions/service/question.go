package service

import (
	repository_assessment "assessment_service/internal/assessments/repository"
	models "assessment_service/internal/model"
	"assessment_service/internal/questions/repository"
	"errors"
)

type QuestionService interface {
	AddQuestion(assessmentID uint, question *models.Question) (*models.Question, error)
	GetQuestionsByAssessment(assessmentID uint) ([]models.Question, error)
	UpdateQuestion(questionID uint, questionData map[string]interface{}) (*models.Question, error)
	DeleteQuestion(questionID uint) error
}

type questionService struct {
	questionRepo   repository.QuestionRepository
	assessmentRepo repository_assessment.AssessmentRepository
}

func NewQuestionService(
	questionRepo repository.QuestionRepository,
	assessmentRepo repository_assessment.AssessmentRepository,
) QuestionService {
	return &questionService{
		questionRepo:   questionRepo,
		assessmentRepo: assessmentRepo,
	}
}

func (s *questionService) AddQuestion(assessmentID uint, question *models.Question) (*models.Question, error) {
	// Check if assessment exists
	_, err := s.assessmentRepo.FindByID(assessmentID)
	if err != nil {
		return nil, errors.New("assessment not found")
	}

	question.AssessmentID = assessmentID

	// Validate question type
	switch question.Type {
	case "multiple-choice":
		// Ensure options are provided
		if len(question.Options) == 0 {
			return nil, errors.New("multiple-choice questions require options")
		}

		// Validate correctAnswer exists in options
		valid := false
		for _, option := range question.Options {
			if option.OptionID == question.CorrectAnswer {
				valid = true
				break
			}
		}

		if !valid {
			return nil, errors.New("correct answer must match one of the option IDs")
		}

	case "true-false":
		// For true-false, the correct answer must be "true" or "false"
		if question.CorrectAnswer != "true" && question.CorrectAnswer != "false" {
			return nil, errors.New("correct answer for true-false questions must be 'true' or 'false'")
		}

	case "essay":
		// Essay questions don't have a correct answer
		question.CorrectAnswer = ""

	default:
		return nil, errors.New("invalid question type")
	}

	// Create question
	err = s.questionRepo.Create(question)
	if err != nil {
		return nil, err
	}

	// Get the created question with options
	return s.questionRepo.FindByID(question.ID)
}

func (s *questionService) GetQuestionsByAssessment(assessmentID uint) ([]models.Question, error) {
	// Check if assessment exists
	_, err := s.assessmentRepo.FindByID(assessmentID)
	if err != nil {
		return nil, errors.New("assessment not found")
	}

	// Get questions
	return s.questionRepo.FindByAssessmentID(assessmentID)
}

func (s *questionService) UpdateQuestion(questionID uint, questionData map[string]interface{}) (*models.Question, error) {
	question, err := s.questionRepo.FindByID(questionID)
	if err != nil {
		return nil, errors.New("question not found")
	}

	// Update text if provided
	if text, ok := questionData["text"].(string); ok {
		question.Text = text
	}

	// Update points if provided
	if points, ok := questionData["points"].(float64); ok {
		question.Points = points
	}

	// Update correct answer if provided
	if correctAnswer, ok := questionData["correctAnswer"]; ok {
		switch question.Type {
		case "multiple-choice":
			// Validate correctAnswer exists in options
			ca, isString := correctAnswer.(string)
			if !isString {
				return nil, errors.New("correct answer must be a string for multiple-choice questions")
			}

			// Validate correctAnswer matches an option
			valid := false
			for _, option := range question.Options {
				if option.OptionID == ca {
					valid = true
					break
				}
			}

			if !valid {
				return nil, errors.New("correct answer must match one of the option IDs")
			}

			question.CorrectAnswer = ca

		case "true-false":
			// Handle both string and boolean
			var isTrueVal bool

			switch v := correctAnswer.(type) {
			case bool:
				isTrueVal = v
			case string:
				isTrueVal = v == "true"
			default:
				return nil, errors.New("correct answer for true-false questions must be a boolean or 'true'/'false' string")
			}

			if isTrueVal {
				question.CorrectAnswer = "true"
			} else {
				question.CorrectAnswer = "false"
			}
		}
	}

	// Update options if provided
	if options, ok := questionData["options"].([]interface{}); ok && question.Type == "multiple-choice" {
		// Delete existing options
		for _, opt := range question.Options {
			err := s.questionRepo.DeleteOption(opt.ID)
			if err != nil {
				return nil, err
			}
		}

		// Create new options
		for _, opt := range options {
			optionMap, ok := opt.(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid options format")
			}

			option := models.QuestionOption{
				QuestionID: question.ID,
				OptionID:   optionMap["id"].(string),
				Text:       optionMap["text"].(string),
			}

			err := s.questionRepo.AddOption(&option)
			if err != nil {
				return nil, err
			}
		}

		// Update question
		err = s.questionRepo.Update(question)
		if err != nil {
			return nil, err
		}
	}
	return s.questionRepo.FindByID(question.ID)
}

func (s *questionService) DeleteQuestion(questionID uint) error {
	// Check if question exists
	_, err := s.questionRepo.FindByID(questionID)
	if err != nil {
		return err
	}

	// Delete question
	return s.questionRepo.Delete(questionID)
}
