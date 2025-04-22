package service

import (
	repository4 "assessment_service/internal/activity/repository"
	repository2 "assessment_service/internal/assessments/repository"
	repository3 "assessment_service/internal/attempts/repository"
	models "assessment_service/internal/model"
	"assessment_service/internal/users/repository"
	"assessment_service/internal/util"
	"go.uber.org/zap"
	"time"
)

type AnalyticsService interface {
	GetUserActivityAnalytics() (map[string]interface{}, error)
	GetAssessmentPerformanceAnalytics() (map[string]interface{}, error)
	ReportActivity(activity *models.Activity) error
	TrackAssessmentSession(sessionData *models.SessionData) error
	LogSuspiciousActivity(activity *models.SuspiciousActivity) error
	GetDashboardSummary() (map[string]interface{}, error)
	GetActivityTimeline() (map[string]interface{}, error)
	GetSystemStatus() (map[string]interface{}, error)
	GetSuspiciousActivity(userID uint, attemptID uint, params util.PaginationParams) ([]models.SuspiciousActivity, int64, error)
}

type analyticsService struct {
	userRepo       repository.UserRepository
	assessmentRepo repository2.AssessmentRepository
	attemptRepo    repository3.AttemptRepository
	activityRepo   repository4.ActivityRepository
	log            *zap.Logger
}

func NewAnalyticsService(
	userRepo repository.UserRepository,
	assessmentRepo repository2.AssessmentRepository,
	attemptRepo repository3.AttemptRepository,
	activityRepo repository4.ActivityRepository,
	log *zap.Logger,
) AnalyticsService {
	return &analyticsService{
		userRepo:       userRepo,
		assessmentRepo: assessmentRepo,
		attemptRepo:    attemptRepo,
		activityRepo:   activityRepo,
		log:            log,
	}
}

func (s *analyticsService) GetUserActivityAnalytics() (map[string]interface{}, error) {
	// Get daily active users for the past week
	dailyActiveUsers, err := s.activityRepo.GetDailyActiveUsers(7)
	if err != nil {
		s.log.Error("[AnalyticsService][GetUserActivityAnalytics] failed to get daily active users", zap.Error(err))
		return nil, err
	}

	// Get activity by hour of day
	activityByHour, err := s.activityRepo.GetActivityByHour()
	if err != nil {
		s.log.Error("[AnalyticsService][GetUserActivityAnalytics] failed to get activity by hour", zap.Error(err))
		return nil, err
	}

	// Get activity by type
	activityByType, err := s.activityRepo.GetActivityByType()
	if err != nil {
		s.log.Error("[AnalyticsService][GetUserActivityAnalytics] failed to get activity by type", zap.Error(err))
		return nil, err
	}

	// Get total active users
	totalActiveUsers, err := s.activityRepo.GetTotalActiveUsers()
	if err != nil {
		s.log.Error("[AnalyticsService][GetUserActivityAnalytics] failed to get total active users", zap.Error(err))
		return nil, err
	}

	// Get new users in last week
	newUsersLastWeek, err := s.userRepo.GetNewUsersCount(7)
	if err != nil {
		s.log.Error("[AnalyticsService][GetUserActivityAnalytics] failed to get new users last week", zap.Error(err))
		return nil, err
	}

	return map[string]interface{}{
		"dailyActiveUsers": dailyActiveUsers,
		"activityByHour":   activityByHour,
		"activityByType":   activityByType,
		"totalActiveUsers": totalActiveUsers,
		"newUsersLastWeek": newUsersLastWeek,
	}, nil
}

func (s *analyticsService) GetAssessmentPerformanceAnalytics() (map[string]interface{}, error) {
	// Get assessment completion rates
	completionRates, err := s.attemptRepo.GetAssessmentCompletionRates()
	if err != nil {
		s.log.Error("[AnalyticsService][GetAssessmentPerformanceAnalytics] failed to get assessment completion rates", zap.Error(err))
		return nil, err
	}

	// Get score distribution
	scoreDistribution, err := s.attemptRepo.GetScoreDistribution()
	if err != nil {
		s.log.Error("[AnalyticsService][GetAssessmentPerformanceAnalytics] failed to get score distribution", zap.Error(err))
		return nil, err
	}

	// Get average time spent
	averageTimeSpent, err := s.attemptRepo.GetAverageTimeSpent()
	if err != nil {
		s.log.Error("[AnalyticsService][GetAssessmentPerformanceAnalytics] failed to get average time spent", zap.Error(err))
		return nil, err
	}

	// Get most challenging assessments
	mostChallenging, err := s.attemptRepo.GetMostChallengingAssessments(2)
	if err != nil {
		s.log.Error("[AnalyticsService][GetAssessmentPerformanceAnalytics] failed to get most challenging assessments", zap.Error(err))
		return nil, err
	}

	// Get most successful assessments
	mostSuccessful, err := s.attemptRepo.GetMostSuccessfulAssessments(2)
	if err != nil {
		s.log.Error("[AnalyticsService][GetAssessmentPerformanceAnalytics] failed to get most successful assessments", zap.Error(err))
		return nil, err
	}

	return map[string]interface{}{
		"assessmentCompletionRates": completionRates,
		"scoreDistribution":         scoreDistribution,
		"averageTimeSpent":          averageTimeSpent,
		"mostChallenging":           mostChallenging,
		"mostSuccessful":            mostSuccessful,
	}, nil
}

func (s *analyticsService) ReportActivity(activity *models.Activity) error {
	// Set timestamp if not provided
	if activity.Timestamp.IsZero() {
		activity.Timestamp = time.Now()
	}

	return s.activityRepo.Create(activity)
}

func (s *analyticsService) TrackAssessmentSession(sessionData *models.SessionData) error {
	// Validate assessment exists
	_, err := s.assessmentRepo.FindByID(sessionData.AssessmentID)
	if err != nil {
		s.log.Error("[AnalyticsService][TrackAssessmentSession] assessment not found", zap.Error(err))
		return err
	}

	// Create activity record
	activity := &models.Activity{
		UserID:       sessionData.UserID,
		Action:       sessionData.Action,
		AssessmentID: &sessionData.AssessmentID,
		Details:      sessionData.Details,
		UserAgent:    sessionData.UserAgent,
		Timestamp:    sessionData.Timestamp,
	}

	return s.activityRepo.Create(activity)
}

func (s *analyticsService) LogSuspiciousActivity(activity *models.SuspiciousActivity) error {
	// Set timestamp if not provided
	if activity.Timestamp.IsZero() {
		activity.Timestamp = time.Now()
	}

	return s.attemptRepo.SaveSuspiciousActivity(activity)
}

func (s *analyticsService) GetDashboardSummary() (map[string]interface{}, error) {
	// Get user stats
	totalUsers, err := s.userRepo.CountAll()
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get total users", zap.Error(err))
		return nil, err
	}

	activeUsers, inactiveUsers, err := s.userRepo.GetUserStats()
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get user stats", zap.Error(err))
		return nil, err
	}

	newThisWeek, err := s.userRepo.GetNewUsersCount(7)
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get new users this week", zap.Error(err))
		return nil, err
	}

	// Get assessment stats
	assessmentStats, err := s.assessmentRepo.GetStatistics()
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get assessment stats", zap.Error(err))
		return nil, err
	}

	// Get attempt stats
	totalAttempts, err := s.attemptRepo.CountAll()
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get total attempts", zap.Error(err))
		return nil, err
	}

	attemptsThisWeek, err := s.attemptRepo.CountByPeriod(7)
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get attempts this week", zap.Error(err))
		return nil, err
	}

	passRate, err := s.attemptRepo.GetPassRate()
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get pass rate", zap.Error(err))
		return nil, err
	}

	// Get users online
	usersOnline, err := s.activityRepo.GetActiveUsers(15) // active in last 15 minutes
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get users online", zap.Error(err))
		return nil, err
	}

	// Get recent suspicious activity
	recentSuspicious, err := s.attemptRepo.CountRecentSuspiciousActivity(24) // last 24 hours
	if err != nil {
		s.log.Error("[AnalyticsService][GetDashboardSummary] failed to get recent suspicious activity", zap.Error(err))
		return nil, err
	}

	return map[string]interface{}{
		"users": map[string]interface{}{
			"total":       totalUsers,
			"active":      activeUsers,
			"inactive":    inactiveUsers,
			"newThisWeek": newThisWeek,
		},
		"assessments": map[string]interface{}{
			"total":       assessmentStats["totalAssessments"],
			"active":      assessmentStats["activeAssessments"],
			"draft":       assessmentStats["draftAssessments"],
			"expired":     assessmentStats["expiredAssessments"],
			"newThisWeek": newThisWeek, // Placeholder - need to implement this in repository
		},
		"activity": map[string]interface{}{
			"assessmentAttempts": map[string]interface{}{
				"total":    totalAttempts,
				"thisWeek": attemptsThisWeek,
				"passRate": passRate,
			},
			"usersOnline":              usersOnline,
			"recentSuspiciousActivity": recentSuspicious,
		},
	}, nil
}

func (s *analyticsService) GetActivityTimeline() (map[string]interface{}, error) {
	// Get recent activity (last 48 hours)
	timeline, err := s.activityRepo.GetRecentActivity(48)
	if err != nil {
		s.log.Error("[AnalyticsService][GetActivityTimeline] failed to get recent activity", zap.Error(err))
		return nil, err
	}

	return map[string]interface{}{
		"timeline": timeline,
	}, nil
}

func (s *analyticsService) GetSystemStatus() (map[string]interface{}, error) {
	// In a real system, we would check database, storage, webcam service, AI service, etc.
	// For this implementation, we'll just return mock data

	return map[string]interface{}{
		"status": "healthy",
		"services": map[string]interface{}{
			"database": "operational",
			"storage":  "operational",
			"webcam":   "operational",
			"ai":       "operational",
		},
		"statistics": map[string]interface{}{
			"uptime":              "15 days, 7 hours",
			"activeConnections":   56,
			"averageResponseTime": 124,  // milliseconds
			"cpuUsage":            35.2, // percentage
			"memoryUsage":         42.8, // percentage
		},
		"lastChecked": time.Now(),
	}, nil
}

func (s *analyticsService) GetSuspiciousActivity(userID uint, attemptID uint, params util.PaginationParams) ([]models.SuspiciousActivity, int64, error) {
	activities, total, err := s.activityRepo.FindSuspiciousActivity(userID, attemptID, params)
	if err != nil {
		s.log.Error("[AnalyticsService][GetSuspiciousActivity] failed to get suspicious activities", zap.Error(err))
		return nil, 0, err
	}

	return activities, total, nil
}
