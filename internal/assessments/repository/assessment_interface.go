package repository

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
)

type AssessmentRepository interface {
	Create(assessment *models.Assessment) error
	FindByID(id uint) (*models.Assessment, error)
	Update(assessment *models.Assessment) error
	Delete(id uint) error
	List(params util.PaginationParams) ([]models.Assessment, int64, error)
	FindRecent(limit int) ([]models.Assessment, error)
	GetStatistics() (map[string]interface{}, error)
	UpdateSettings(id uint, settings *models.AssessmentSettings) error
	GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error)
	Publish(id uint) error
	Duplicate(assessment *models.Assessment) error
	GetAssessmentHasAttemptByUser(params util.PaginationParams, userID uint) ([]models.Assessment, int64, error)
}
