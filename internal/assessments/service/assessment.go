package service

import (
	"assessment_service/internal/assessments/repository"
	models "assessment_service/internal/model"
	repository2 "assessment_service/internal/users/repository"
	"assessment_service/internal/util"
	"errors"
	"go.uber.org/zap"
	"time"
)

type AssessmentService interface {
	Create(assessment *models.Assessment) error
	GetByID(id uint) (*models.Assessment, error)
	Update(id uint, assessmentData map[string]interface{}) (*models.Assessment, error)
	Delete(id uint) error
	List(params util.PaginationParams) ([]models.Assessment, int64, error)
	GetRecentAssessments(limit int) ([]models.Assessment, error)
	GetStatistics() (map[string]interface{}, error)
	UpdateSettings(id uint, settings *models.AssessmentSettings) error
	GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error)
	Publish(id uint) (*models.Assessment, error)
	Duplicate(id uint, newTitle string, copyQuestions, copySettings, setAsDraft bool) (*models.Assessment, error)

	GetAssessmentDetailWithUser(assessmentID uint, params util.PaginationParams) (*models.Assessment, []models.User, int64, error)
}

type assessmentService struct {
	assessmentRepo repository.AssessmentRepository
	log            *zap.Logger
	userRepo       repository2.UserRepository
}

func NewAssessmentService(
	assessmentRepo repository.AssessmentRepository,
) AssessmentService {
	return &assessmentService{
		assessmentRepo: assessmentRepo,
	}
}

func (s *assessmentService) Create(assessment *models.Assessment) error {
	// Set default status
	if assessment.Status == "" {
		assessment.Status = "Draft"
	}

	// Create default settings
	assessment.Settings = models.AssessmentSettings{
		RandomizeQuestions:          false,
		ShowResults:                 true,
		AllowRetake:                 false,
		MaxAttempts:                 1,
		TimeLimitEnforced:           true,
		RequireWebcam:               false,
		PreventTabSwitching:         false,
		RequireIdentityVerification: false,
	}

	return s.assessmentRepo.Create(assessment)
}

func (s *assessmentService) GetByID(id uint) (*models.Assessment, error) {
	return s.assessmentRepo.FindByID(id)
}

func (s *assessmentService) Update(id uint, assessmentData map[string]interface{}) (*models.Assessment, error) {
	assessment, err := s.assessmentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if title, ok := assessmentData["title"].(string); ok {
		assessment.Title = title
	}

	if subject, ok := assessmentData["subject"].(string); ok {
		assessment.Subject = subject
	}

	if description, ok := assessmentData["description"].(string); ok {
		assessment.Description = description
	}

	if duration, ok := assessmentData["duration"].(float64); ok {
		assessment.Duration = int(duration)
	}

	if status, ok := assessmentData["status"].(string); ok {
		assessment.Status = status
	}

	if dueDateStr, ok := assessmentData["dueDate"].(string); ok && dueDateStr != "" {
		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err == nil {
			assessment.DueDate = &dueDate
		}
	}

	if passingScore, ok := assessmentData["passingScore"].(float64); ok {
		assessment.PassingScore = passingScore
	}

	// Update assessment
	err = s.assessmentRepo.Update(assessment)
	if err != nil {
		return nil, err
	}

	return assessment, nil
}

func (s *assessmentService) Delete(id uint) error {
	// Check if assessment exists
	_, err := s.assessmentRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Delete assessment
	return s.assessmentRepo.Delete(id)
}

func (s *assessmentService) List(params util.PaginationParams) ([]models.Assessment, int64, error) {
	return s.assessmentRepo.List(params)
}

func (s *assessmentService) GetRecentAssessments(limit int) ([]models.Assessment, error) {
	return s.assessmentRepo.FindRecent(limit)
}

func (s *assessmentService) GetStatistics() (map[string]interface{}, error) {
	return s.assessmentRepo.GetStatistics()
}

func (s *assessmentService) UpdateSettings(id uint, settings *models.AssessmentSettings) error {
	// Check if assessment exists
	_, err := s.assessmentRepo.FindByID(id)
	if err != nil {
		return err
	}

	return s.assessmentRepo.UpdateSettings(id, settings)
}

func (s *assessmentService) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	return s.assessmentRepo.GetResults(id, params)
}

func (s *assessmentService) Publish(id uint) (*models.Assessment, error) {
	// Check if assessment exists
	assessment, err := s.assessmentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if assessment has questions
	if len(assessment.Questions) == 0 {
		return nil, errors.New("cannot publish assessment without questions")
	}

	// Update status
	err = s.assessmentRepo.Publish(id)
	if err != nil {
		return nil, err
	}

	// Refresh the assessment
	return s.assessmentRepo.FindByID(id)
}

func (s *assessmentService) Duplicate(id uint, newTitle string, copyQuestions, copySettings, setAsDraft bool) (*models.Assessment, error) {
	// Check if assessment exists
	originalAssessment, err := s.assessmentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// If we don't want to copy questions or settings, set them to empty
	if !copyQuestions {
		originalAssessment.Questions = []models.Question{}
	}

	if !copySettings {
		originalAssessment.Settings = models.AssessmentSettings{}
	}

	// Set new title and status
	if newTitle != "" {
		originalAssessment.Title = newTitle
	} else {
		originalAssessment.Title = originalAssessment.Title + " (Copy)"
	}

	if setAsDraft {
		originalAssessment.Status = "Draft"
	}

	// Create duplicate
	err = s.assessmentRepo.Duplicate(originalAssessment)
	if err != nil {
		return nil, err
	}

	// Get the newly created assessment (would be the most recent one with this title)
	assessments, _, err := s.assessmentRepo.List(util.PaginationParams{
		Limit: 1,
		Filters: map[string]interface{}{
			"title": originalAssessment.Title,
		},
		SortBy:  "created_at",
		SortDir: "DESC",
	})

	if err != nil || len(assessments) == 0 {
		return nil, errors.New("failed to retrieve the duplicated assessment")
	}

	return &assessments[0], nil
}

func (s *assessmentService) GetAssessmentDetailWithUser(assessmentID uint, params util.PaginationParams) (*models.Assessment, []models.User, int64, error) {
	// Get the attempt by ID
	attempt, err := s.assessmentRepo.FindByID(assessmentID)
	if err != nil {
		return nil, nil, 0, err
	}

	// if params is nil, set default values
	if params.Limit == 0 {
		params.Limit = 10
	}

	// Get the user details for the attempt
	users, total, err := s.userRepo.GetListUserByAttempt(params, assessmentID)
	if err != nil {
		return nil, nil, 0, err
	}

	return attempt, users, total, nil
}
