package postgres

import (
	"assessment_service/internal/assessments/repository"
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"errors"

	"gorm.io/gorm"
)

type assessmentRepository struct {
	db *gorm.DB
}

func (a assessmentRepository) Create(assessment *models.Assessment) error {
	return a.db.Create(assessment).Error
}

func (a assessmentRepository) FindByID(id uint) (*models.Assessment, error) {
	var assessment models.Assessment
	err := a.db.Preload("Questions.Options").
		Preload("Settings").
		Preload("CreatedBy").First(&assessment, id).Error
	if err != nil {
		return nil, err
	}
	return &assessment, nil
}

func (a assessmentRepository) Update(assessment *models.Assessment) error {
	return a.db.Save(assessment).Error
}

func (a assessmentRepository) Delete(id uint) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		// delete question
		if err := tx.Where("assessment_id = ?", id).Delete(&models.Question{}).Error; err != nil {
			return err
		}

		// delete settings
		if err := tx.Where("assessment_id = ?", id).Delete(&models.AssessmentSettings{}).Error; err != nil {
			return err
		}

		// Delete the assessment
		if err := tx.Delete(&models.Assessment{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (a assessmentRepository) List(params util.PaginationParams) ([]models.Assessment, int64, error) {
	var assessments []models.Assessment
	var count int64

	query := a.db.Model(&models.Assessment{})

	if params.Search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if params.Filters != nil {
		if val, ok := params.Filters["status"]; ok {
			query = query.Where("status = ?", val)
		}

		if val, ok := params.Filters["status"]; ok {
			query = query.Where("status = ?", val)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortDir)
	} else {
		query = query.Order("created_at DESC")
	}

	query = query.Offset(params.Offset).Limit(params.Limit)

	query = query.
		Preload("CreatedBy").
		Select("assessments.*, COUNT(questions.id) as questions_count").
		Joins("LEFT JOIN questions ON questions.assessment_id = assessments.id").
		Group("assessments.id")

	if err := query.Find(&assessments).Error; err != nil {
		return nil, 0, err
	}

	return assessments, count, nil
}

func (a assessmentRepository) FindRecent(limit int) ([]models.Assessment, error) {
	var assessments []models.Assessment

	err := a.db.
		Preload("CreatedBy").
		Order("created_at DESC").
		Limit(limit).
		Find(&assessments).Error

	if err != nil {
		return nil, err
	}

	return assessments, nil
}

func (a assessmentRepository) GetStatistics() (map[string]interface{}, error) {
	var totalAssessments, activeAssessments, draftAssessments, expiredAssessments int64
	var totalAttempts int64
	// var averageScore float64

	// Count assessments by status
	if err := a.db.Model(&models.Assessment{}).Count(&totalAssessments).Error; err != nil {
		return nil, err
	}

	if err := a.db.Model(&models.Assessment{}).Where("status = ?", "Active").Count(&activeAssessments).Error; err != nil {
		return nil, err
	}

	if err := a.db.Model(&models.Assessment{}).Where("status = ?", "Draft").Count(&draftAssessments).Error; err != nil {
		return nil, err
	}

	if err := a.db.Model(&models.Assessment{}).Where("status = ?", "Expired").Count(&expiredAssessments).Error; err != nil {
		return nil, err
	}

	// Count all attempts
	if err := a.db.Model(&models.Attempt{}).Count(&totalAttempts).Error; err != nil {
		return nil, err
	}

	// Calculate average score and pass rate
	var scoreResult struct {
		AvgScore float64
		PassRate float64
	}

	err := a.db.Model(&models.Attempt{}).
		Select("AVG(score) as avg_score, SUM(CASE WHEN status = 'Passed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as pass_rate").
		Where("score IS NOT NULL").
		Scan(&scoreResult).Error

	if err != nil {
		return nil, err
	}

	// Get counts by subject
	type SubjectCount struct {
		Subject string
		Count   int
	}

	var subjectCounts []SubjectCount
	err = a.db.Model(&models.Assessment{}).
		Select("subject, COUNT(*) as count").
		Group("subject").
		Order("count DESC").
		Scan(&subjectCounts).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalAssessments":   totalAssessments,
		"activeAssessments":  activeAssessments,
		"draftAssessments":   draftAssessments,
		"expiredAssessments": expiredAssessments,
		"totalAttempts":      totalAttempts,
		"passRate":           scoreResult.PassRate,
		"averageScore":       scoreResult.AvgScore,
		"bySubject":          subjectCounts,
	}, nil
}

func (a assessmentRepository) UpdateSettings(id uint, settings *models.AssessmentSettings) error {

	var currentSettings models.AssessmentSettings
	if err := a.db.Where("assessment_id = ?", id).First(&currentSettings).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			settings.ID = id
			return a.db.Create(&settings).Error
		}
		return err
	}

	return a.db.Model(&settings).Updates(settings).Error
}

func (a assessmentRepository) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {

	return nil, 0, nil
}

func (a assessmentRepository) Publish(id uint) error {
	//TODO implement me
	panic("implement me")
}

func (a assessmentRepository) Duplicate(assessment *models.Assessment) error {
	return a.db.Transaction(func(tx *gorm.DB) error {

		return nil
	})
}

func newAssessmentRepository(db *gorm.DB) repository.AssessmentRepository {
	return &assessmentRepository{db: db}
}
