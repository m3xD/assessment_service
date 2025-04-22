package service

import (
	repository2 "assessment_service/internal/attempts/repository"
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"go.uber.org/zap"
)

type AttemptService interface {
	GetListAttemptByUserAndAssessment(userID, assessmentID uint, params util.PaginationParams) ([]models.Attempt, int64, error)
	GetAttemptDetail(attemptID uint) (*models.Attempt, error)
	GradeAttempt(newAttempt models.Attempt) error
}

type attemptService struct {
	attemptRepo repository2.AttemptRepository
	log         *zap.Logger
}

func NewAttemptService(
	attemptRepo repository2.AttemptRepository,
	log *zap.Logger,
) AttemptService {
	return &attemptService{
		attemptRepo: attemptRepo,
		log:         log,
	}
}

func (s *attemptService) GetListAttemptByUserAndAssessment(userID, assessmentID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	attempts, total, err := s.attemptRepo.ListAttemptByUserAndAssessmentID(userID, assessmentID, params)

	if err != nil {
		s.log.Error("Failed to get attempts", zap.Error(err))
		return nil, 0, err
	}

	return attempts, total, nil
}

func (s *attemptService) GetAttemptDetail(attemptID uint) (*models.Attempt, error) {
	attempt, err := s.attemptRepo.FindByID(attemptID)
	if err != nil {
		s.log.Error("Failed to get attempt detail", zap.Error(err))
		return nil, err
	}

	return attempt, nil
}

func (s *attemptService) GradeAttempt(newAttempt models.Attempt) error {
	err := s.attemptRepo.Update(&newAttempt)
	if err != nil {
		s.log.Error("[GradeAttempt] Failed to update attempt grade", zap.Error(err))
		return err
	}

	return nil
}
