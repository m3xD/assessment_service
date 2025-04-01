package repository

import (
	models "assessment_service/internal/model"
	"gorm.io/gorm"
)

type QuestionRepository interface {
	Create(question *models.Question) error
	FindByID(id uint) (*models.Question, error)
	FindByAssessmentID(assessmentID uint) ([]models.Question, error)
	Update(question *models.Question) error
	Delete(id uint) error
	AddOption(option *models.QuestionOption) error
	UpdateOption(option *models.QuestionOption) error
	DeleteOption(id uint) error
}

type questionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) QuestionRepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) Create(question *models.Question) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		temp := question.Options
		question.Options = nil
		// Create question
		if err := tx.Create(question).Error; err != nil {
			return err
		}

		if question.Type != "essay" {
			question.Options = temp
			// Create options if any
			for i := range question.Options {
				question.Options[i].QuestionID = question.ID
			}

			if err := tx.Create(question.Options).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *questionRepository) FindByID(id uint) (*models.Question, error) {
	var question models.Question
	err := r.db.
		Preload("Options").
		First(&question, id).Error

	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *questionRepository) FindByAssessmentID(assessmentID uint) ([]models.Question, error) {
	var questions []models.Question
	err := r.db.
		Where("assessment_id = ?", assessmentID).
		Preload("Options").
		Order("id").
		Find(&questions).Error

	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *questionRepository) Update(question *models.Question) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update question
		if err := tx.Save(question).Error; err != nil {
			return err
		}

		// Handle options separately in caller

		return nil
	})
}

func (r *questionRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete options first
		if err := tx.Where("question_id = ?", id).Delete(&models.QuestionOption{}).Error; err != nil {
			return err
		}

		// Delete question
		return tx.Delete(&models.Question{}, id).Error
	})
}

func (r *questionRepository) AddOption(option *models.QuestionOption) error {
	return r.db.Create(option).Error
}

func (r *questionRepository) UpdateOption(option *models.QuestionOption) error {
	return r.db.Save(option).Error
}

func (r *questionRepository) DeleteOption(id uint) error {
	return r.db.Delete(&models.QuestionOption{}, id).Error
}
