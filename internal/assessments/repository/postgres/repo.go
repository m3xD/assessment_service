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

	currentSettings.RandomizeQuestions = settings.RandomizeQuestions
	currentSettings.ShowResults = settings.ShowResults
	currentSettings.AllowRetake = settings.AllowRetake
	currentSettings.MaxAttempts = settings.MaxAttempts
	currentSettings.TimeLimitEnforced = settings.TimeLimitEnforced
	currentSettings.RequireWebcam = settings.RequireWebcam
	currentSettings.PreventTabSwitching = settings.PreventTabSwitching
	currentSettings.RequireIdentityVerification = settings.RequireIdentityVerification

	return a.db.Save(&currentSettings).Error
}

func (a assessmentRepository) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	var result []map[string]interface{}
	var total int64

	// check if assessment is exist
	var assessment models.Assessment
	if err := a.db.First(&assessment).Error; err != nil {
		return nil, 0, err
	}

	query := a.db.Model(&models.Attempt{}).Joins("JOIN users u ON u.id = attempts.user_id").
		Select(`attempts.id, 
					users.name as user, 
					attempts.user_id as user_id,
					DATE_FORMAT(attempts.submitted_at, '%Y-%m-%d') as date,
					DATE_FORMAT(attempts.submitted_at, '%H:%i') as time,
					attempts.score, 
					attempts.duration, 
					attempts.status,
					attempts.submitted_at`).
		Where("attempts.assessment_id = ? AND attempts.submitted_at IS NOT NULL", id)

	// apply filter
	if params.Filters != nil {
		if val, ok := params.Filters["user"].(string); ok && val != "" {
			query = query.Where("users.name LIKE ? OR users.email LIKE ?", "%"+val+"%", "%"+val+"%")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// apply sorting and pagination
	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortDir)
	} else {
		query = query.Order("attempts.submitted_at DESC")
	}

	// apply offset and limit
	query = query.Offset(params.Offset).Limit(params.Limit)

	if err := a.db.Table("attempts").Find(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (a assessmentRepository) Publish(id uint) error {
	return a.db.Model(&models.Assessment{}).Where("id = ?", id).Update("status", "Active").Error
}

func (a assessmentRepository) Duplicate(assessment *models.Assessment) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		// Create a copy of the assessment
		assessmentCopy := models.Assessment{
			Title:        assessment.Title,
			Subject:      assessment.Subject,
			Description:  assessment.Description,
			Duration:     assessment.Duration,
			Status:       assessment.Status,
			DueDate:      assessment.DueDate,
			CreatedByID:  assessment.CreatedByID,
			PassingScore: assessment.PassingScore,
		}

		if err := tx.Create(&assessmentCopy).Error; err != nil {
			return err
		}

		// Copy the settings if they exist
		if assessment.Settings.ID != 0 {
			settingsCopy := models.AssessmentSettings{
				AssessmentID:                assessmentCopy.ID,
				RandomizeQuestions:          assessment.Settings.RandomizeQuestions,
				ShowResults:                 assessment.Settings.ShowResults,
				AllowRetake:                 assessment.Settings.AllowRetake,
				MaxAttempts:                 assessment.Settings.MaxAttempts,
				TimeLimitEnforced:           assessment.Settings.TimeLimitEnforced,
				RequireWebcam:               assessment.Settings.RequireWebcam,
				PreventTabSwitching:         assessment.Settings.PreventTabSwitching,
				RequireIdentityVerification: assessment.Settings.RequireIdentityVerification,
			}

			if err := tx.Create(&settingsCopy).Error; err != nil {
				return err
			}
		}

		// Copy the questions if needed
		for _, question := range assessment.Questions {
			questionCopy := models.Question{
				AssessmentID:  assessmentCopy.ID,
				Type:          question.Type,
				Text:          question.Text,
				CorrectAnswer: question.CorrectAnswer,
				Points:        question.Points,
			}

			if err := tx.Create(&questionCopy).Error; err != nil {
				return err
			}

			// Copy the options for multiple-choice questions
			for _, option := range question.Options {
				optionCopy := models.QuestionOption{
					QuestionID: questionCopy.ID,
					OptionID:   option.OptionID,
					Text:       option.Text,
				}

				if err := tx.Create(&optionCopy).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (a assessmentRepository) GetAssessmentHasAttemptByUser(params util.PaginationParams, userID uint) ([]models.Assessment, int64, error) {
	var assessments []models.Assessment
	var total int64

	query := a.db.Model(&models.Assessment{}).
		Joins("JOIN attempts at ON at.assessment_id = assessments.id").
		Where("at.user_id = ?", userID)

	// Apply filters
	if params.Search != "" {
		query = query.Where("assessments.title LIKE ?", "%"+params.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortDir)
	} else {
		query = query.Order("assessments.created_at DESC")
	}

	query = query.Offset(params.Offset).Limit(params.Limit)

	if err := query.Find(&assessments).Error; err != nil {
		return nil, 0, err
	}

	return assessments, total, nil
}

func NewAssessmentRepository(db *gorm.DB) repository.AssessmentRepository {
	return &assessmentRepository{db: db}
}
