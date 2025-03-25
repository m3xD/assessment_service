package service

import (
	repository4 "assessment_service/internal/activity/repository"
	repository2 "assessment_service/internal/assessments/repository"
	repository3 "assessment_service/internal/attempts/repository"
	models "assessment_service/internal/model"
	"assessment_service/internal/users/repository"
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
}

type analyticsService struct {
	userRepo       repository.UserRepository
	assessmentRepo repository2.AssessmentRepository
	attemptRepo    repository3.AttemptRepository
	activityRepo   repository4.ActivityRepository
}

func NewAnalyticsService(
	userRepo repository.UserRepository,
	assessmentRepo repository2.AssessmentRepository,
	attemptRepo repository3.AttemptRepository,
	activityRepo repository4.ActivityRepository,
) AnalyticsService {
	return &analyticsService{
		userRepo:       userRepo,
		assessmentRepo: assessmentRepo,
		attemptRepo:    attemptRepo,
		activityRepo:   activityRepo,
	}
}

func (s *analyticsService) GetUserActivityAnalytics() (map[string]interface{}, error) {
	// Get daily active users for the past week
	dailyActiveUsers, err := s.activityRepo.GetDailyActiveUsers(7)
	if err != nil {
		return nil, err
	}

	// Get activity by hour of day
	activityByHour, err := s.activityRepo.GetActivityByHour()
	if err != nil {
		return nil, err
	}

	// Get activity by type
	activityByType, err := s.activityRepo.GetActivityByType()
	if err != nil {
		return nil, err
	}

	// Get total active users
	totalActiveUsers, err := s.activityRepo.GetTotalActiveUsers()
	if err != nil {
		return nil, err
	}

	// Get new users in last week
	newUsersLastWeek, err := s.userRepo.GetNewUsersCount(7)
	if err != nil {
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
		return nil, err
	}

	// Get score distribution
	scoreDistribution, err := s.attemptRepo.GetScoreDistribution()
	if err != nil {
		return nil, err
	}

	// Get average time spent
	averageTimeSpent, err := s.attemptRepo.GetAverageTimeSpent()
	if err != nil {
		return nil, err
	}

	// Get most challenging assessments
	mostChallenging, err := s.attemptRepo.GetMostChallengingAssessments(2)
	if err != nil {
		return nil, err
	}

	// Get most successful assessments
	mostSuccessful, err := s.attemptRepo.GetMostSuccessfulAssessments(2)
	if err != nil {
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
		return nil, err
	}

	activeUsers, inactiveUsers, err := s.userRepo.GetUserStats()
	if err != nil {
		return nil, err
	}

	newThisWeek, err := s.userRepo.GetNewUsersCount(7)
	if err != nil {
		return nil, err
	}

	// Get assessment stats
	assessmentStats, err := s.assessmentRepo.GetStatistics()
	if err != nil {
		return nil, err
	}

	// Get attempt stats
	totalAttempts, err := s.attemptRepo.CountAll()
	if err != nil {
		return nil, err
	}

	attemptsThisWeek, err := s.attemptRepo.CountByPeriod(7)
	if err != nil {
		return nil, err
	}

	passRate, err := s.attemptRepo.GetPassRate()
	if err != nil {
		return nil, err
	}

	// Get users online
	usersOnline, err := s.activityRepo.GetActiveUsers(15) // active in last 15 minutes
	if err != nil {
		return nil, err
	}

	// Get recent suspicious activity
	recentSuspicious, err := s.attemptRepo.CountRecentSuspiciousActivity(24) // last 24 hours
	if err != nil {
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
