package repository

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AttemptRepository defines operations for managing assessment attempts
type AttemptRepository interface {
	// Core attempt operations
	Create(attempt *models.Attempt) error
	FindByID(id uint) (*models.Attempt, error)
	Update(attempt *models.Attempt) error
	Delete(id uint) error

	// Answer management
	SaveAnswer(answer *models.Answer) error
	UpdateAnswer(answer *models.Answer) error
	FindAnswersByAttemptID(attemptID uint) ([]models.Answer, error)
	FindAnswerByAttemptAndQuestion(attemptID, questionID uint) (*models.Answer, error)

	// Student assessment interactions
	FindAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error)
	HasCompletedAssessment(userID, assessmentID uint) (bool, error)
	CountAttemptsByUserAndAssessment(userID, assessmentID uint) (int, error)
	FindCompletedAttemptsByUserAndAssessment(userID, assessmentID uint) ([]map[string]interface{}, error)

	// Statistics and analytics
	GetAssessmentCompletionRates() (map[string]interface{}, error)
	GetScoreDistribution() (map[string]interface{}, error)
	GetAverageTimeSpent() (map[string]interface{}, error)
	GetMostChallengingAssessments(limit int) ([]map[string]interface{}, error)
	GetMostSuccessfulAssessments(limit int) ([]map[string]interface{}, error)
	GetPassRate() (float64, error)

	// General statistics
	CountAll() (int64, error)
	CountByPeriod(days int) (int64, error)

	// Proctoring and monitoring
	SaveSuspiciousActivity(activity *models.SuspiciousActivity) error
	CountRecentSuspiciousActivity(hours int) (int64, error)
	FindSuspiciousActivitiesByAttemptID(attemptID uint) ([]models.SuspiciousActivity, error)
}

type attemptRepository struct {
	db *gorm.DB
}

// NewAttemptRepository creates a new instance of AttemptRepository
func NewAttemptRepository(db *gorm.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

// Create inserts a new attempt record into the database
func (r *attemptRepository) Create(attempt *models.Attempt) error {
	now := time.Now()
	attempt.CreatedAt = now
	attempt.UpdatedAt = now

	result := r.db.Create(&attempt)
	if result.Error != nil {
		return fmt.Errorf("failed to create attempt: %w", result.Error)
	}

	return nil
}

// FindByID finds an attempt by its ID
func (r *attemptRepository) FindByID(id uint) (*models.Attempt, error) {
	var attempt models.Attempt

	result := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&attempt)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("attempt with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to find attempt: %w", result.Error)
	}

	// Load answers for this attempt
	var answers []models.Answer
	if err := r.db.Where("attempt_id = ?", id).Find(&answers).Error; err != nil {
		return nil, fmt.Errorf("failed to load answers: %w", err)
	}
	attempt.Answers = answers

	return &attempt, nil
}

// Update updates an attempt record in the database
func (r *attemptRepository) Update(attempt *models.Attempt) error {
	attempt.UpdatedAt = time.Now()

	result := r.db.Save(&attempt)
	if result.Error != nil {
		return fmt.Errorf("failed to update attempt: %w", result.Error)
	}

	return nil
}

// Delete soft-deletes an attempt by setting its deleted_at field
func (r *attemptRepository) Delete(id uint) error {
	result := r.db.Model(&models.Attempt{}).Where("id = ?", id).Update("deleted_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to delete attempt: %w", result.Error)
	}

	return nil
}

// SaveAnswer inserts a new answer record into the database
func (r *attemptRepository) SaveAnswer(answer *models.Answer) error {
	now := time.Now()
	answer.CreatedAt = now
	answer.UpdatedAt = now

	result := r.db.Create(&answer)
	if result.Error != nil {
		return fmt.Errorf("failed to save answer: %w", result.Error)
	}

	return nil
}

// UpdateAnswer updates an existing answer in the database
func (r *attemptRepository) UpdateAnswer(answer *models.Answer) error {
	answer.UpdatedAt = time.Now()

	result := r.db.Save(&answer)
	if result.Error != nil {
		return fmt.Errorf("failed to update answer: %w", result.Error)
	}

	return nil
}

// FindAnswersByAttemptID retrieves all answers for a given attempt
func (r *attemptRepository) FindAnswersByAttemptID(attemptID uint) ([]models.Answer, error) {
	var answers []models.Answer

	result := r.db.Where("attempt_id = ?", attemptID).Find(&answers)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find answers: %w", result.Error)
	}

	return answers, nil
}

// FindAnswerByAttemptAndQuestion finds an answer for a specific question in an attempt
func (r *attemptRepository) FindAnswerByAttemptAndQuestion(attemptID, questionID uint) (*models.Answer, error) {
	var answer models.Answer

	result := r.db.Where("attempt_id = ? AND question_id = ?", attemptID, questionID).First(&answer)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find answer: %w", result.Error)
	}

	return &answer, nil
}

// FindAvailableAssessments finds assessments that are available for a user to take
func (r *attemptRepository) FindAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	var total int64
	var results []map[string]interface{}

	// Subquery for counting user attempts
	attemptCountSubquery := r.db.Model(&models.Attempt{}).
		Select("COUNT(*)").
		Where("user_id = ? AND assessment_id = assessments.id", userID)

	// Base query
	query := r.db.Table("assessments").
		Select("assessments.id, assessments.title, assessments.description, assessments.subject, "+
			"assessments.duration, assessments.passing_score, assessments.due_date, "+
			"assessments.created_at, users.name AS creator_name, assessment_settings.randomize_questions, "+
			"assessment_settings.show_results, assessment_settings.allow_retake, assessment_settings.max_attempts, "+
			"assessment_settings.time_limit_enforced, assessment_settings.require_webcam, assessment_settings.prevent_tab_switching, "+
			"assessment_settings.require_identity_verification, "+
			"(?) AS attempt_count", attemptCountSubquery).
		Joins("JOIN users ON assessments.created_by_id = users.id").
		Joins("LEFT JOIN assessment_settings ON assessments.id = assessment_settings.assessment_id").
		Where("assessments.status = ? AND assessments.created_at <= ? AND (assessments.due_date IS NULL OR assessments.due_date >= ?)",
			"Active", time.Now(), time.Now())

	// Apply search filter if provided
	if search, ok := params.Filters["search"].(string); ok && search != "" {
		query = query.Where("assessments.title LIKE ? OR assessments.description LIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	// Apply subject filter if provided
	if subject, ok := params.Filters["subject"].(string); ok && subject != "" {
		query = query.Where("assessments.subject = ?", subject)
	}

	// Get total count
	countQuery := query
	countQuery.Count(&total)

	// Apply sorting
	if params.SortBy != "" {
		query = query.Order(fmt.Sprintf("%s %s", params.SortBy, params.SortDir))
	} else {
		query = query.Order("assessments.created_at DESC")
	}

	// Apply pagination
	query = query.Limit(int(params.Limit)).Offset(int(params.Offset))

	// Execute the query
	rows, err := query.Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query available assessments: %w", err)
	}
	defer rows.Close()

	// Scan results into maps
	for rows.Next() {
		var id, creatorName, title, description, subject string
		var duration int
		var passingScore float64
		var dueDate, createdAt time.Time
		var shuffleQuestions, showResults, allowRetake, timeLimitEnforcer, requireWebcam, preventTabSwitching, requireIdentifyVerification bool
		var maxAttempts, attemptCount int

		err := rows.Scan(
			&id, &title, &description, &subject, &duration, &passingScore,
			&dueDate, &createdAt, &creatorName, &shuffleQuestions,
			&showResults, &allowRetake, &maxAttempts, &timeLimitEnforcer, &requireWebcam, &preventTabSwitching, &requireIdentifyVerification, &attemptCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning assessment row: %w", err)
		}

		// Create a result map
		result := map[string]interface{}{
			"id":                            id,
			"title":                         title,
			"description":                   description,
			"subject":                       subject,
			"duration":                      duration,
			"passing_score":                 passingScore,
			"due_date":                      dueDate,
			"created_at":                    createdAt,
			"creator_name":                  creatorName,
			"shuffle_questions":             shuffleQuestions,
			"show_results":                  showResults,
			"allow_retake":                  allowRetake,
			"time_limit_enforcer":           timeLimitEnforcer,
			"require_webcam":                requireWebcam,
			"prevent_tab_switching":         preventTabSwitching,
			"require_identity_verification": requireIdentifyVerification,
			"max_attempts":                  maxAttempts,
			"attempt_count":                 attemptCount,
			"can_attempt":                   maxAttempts == 0 || attemptCount < maxAttempts,
		}

		results = append(results, result)
	}

	return results, total, nil
}

// HasCompletedAssessment checks if a user has completed an assessment
func (r *attemptRepository) HasCompletedAssessment(userID, assessmentID uint) (bool, error) {
	var count int64

	result := r.db.Model(&models.Attempt{}).
		Where("user_id = ? AND assessment_id = ? AND status = ?", userID, assessmentID, "completed").
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check completion status: %w", result.Error)
	}

	return count > 0, nil
}

// CountAttemptsByUserAndAssessment counts a user's attempts for a specific assessment
func (r *attemptRepository) CountAttemptsByUserAndAssessment(userID, assessmentID uint) (int, error) {
	var count int64

	result := r.db.Model(&models.Attempt{}).
		Where("user_id = ? AND assessment_id = ?", userID, assessmentID).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to count attempts: %w", result.Error)
	}

	return int(count), nil
}

// FindCompletedAttemptsByUserAndAssessment finds all completed attempts by a user for a specific assessment
func (r *attemptRepository) FindCompletedAttemptsByUserAndAssessment(userID, assessmentID uint) ([]map[string]interface{}, error) {
	type Result struct {
		AssessmentID uint      `gorm:"column:assessment_id" json:"assessmentId"`
		ID           uint      `gorm:"column:id" json:"attemptId"`
		StartedAt    time.Time `gorm:"column:started_at" json:"startedAt"`
		SubmittedAt  time.Time `gorm:"column:submitted_at" json:"submittedAt"`
		Score        float64   `gorm:"column:score"`
		Duration     int       `gorm:"column:duration"`
		Status       string    `gorm:"column:status"`
		Title        string    `gorm:"column:title"`
		PassingScore float64   `gorm:"column:passing_score" json:"passingScore"`
		Feedback     string    `gorm:"column:feedback" json:"feedback"`
	}

	var results []Result

	err := r.db.Table("attempts").
		Select("attempts.id, attempts.assessment_id, attempts.started_at, attempts.submitted_at, attempts.score, attempts.duration, attempts.status, assessments.title, assessments.passing_score, attempts.feedback").
		Joins("JOIN assessments ON attempts.assessment_id = assessments.id").
		Where("attempts.user_id = ? AND attempts.assessment_id = ? AND attempts.deleted_at IS NULL", userID, assessmentID).
		Order("attempts.submitted_at DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query attempts: %w", err)
	}

	// Convert to map[string]interface{}
	resultMaps := make([]map[string]interface{}, len(results))
	for i, r := range results {
		resultMaps[i] = map[string]interface{}{
			"assessmentId": r.AssessmentID,
			"attemptId":    r.ID,
			"startedAt":    r.StartedAt,
			"submittedAt":  r.SubmittedAt,
			"score":        r.Score,
			"duration":     r.Duration,
			"status":       r.Status,
			"title":        r.Title,
			"passingScore": r.PassingScore,
			// "passed":       r.Score >= r.PassingScore,
			"feedback": r.Feedback,
		}
	}

	return resultMaps, nil
}

// GetAssessmentCompletionRates calculates completion rates for assessments
func (r *attemptRepository) GetAssessmentCompletionRates() (map[string]interface{}, error) {
	type CompletionRate struct {
		ID                uint    `gorm:"column:id"`
		Title             string  `gorm:"column:title"`
		TotalAttempts     int64   `gorm:"column:total_attempts"`
		CompletedAttempts int64   `gorm:"column:completed_attempts"`
		TotalUsers        int64   `gorm:"column:total_users"`
		CompletionRate    float64 `gorm:"column:completion_rate"`
	}

	var results []CompletionRate

	// This is a complex query that requires raw SQL in GORM
	err := r.db.Raw(`
		WITH assessment_stats AS (
			SELECT 
				a.id, a.title,
				COUNT(DISTINCT att.id) as total_attempts,
				COUNT(DISTINCT CASE WHEN att.status = 'Completed' THEN att.id END) as completed_attempts,
				COUNT(DISTINCT att.user_id) as total_users
			FROM assessments a
			LEFT JOIN attempts att ON a.id = att.assessment_id
			GROUP BY a.id, a.title
		)
		SELECT 
			id, title, total_attempts, completed_attempts, total_users,
			CASE WHEN total_attempts > 0 THEN 
				ROUND((completed_attempts::numeric / total_attempts::numeric) * 100, 2)
			ELSE
				0
			END as completion_rate
		FROM assessment_stats
		ORDER BY total_attempts DESC
		LIMIT 10
	`).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch assessment completion rates: %w", err)
	}

	// Convert to map[string]interface{}
	resultMaps := make([]map[string]interface{}, len(results))
	for i, r := range results {
		resultMaps[i] = map[string]interface{}{
			"id":                 r.ID,
			"title":              r.Title,
			"total_attempts":     r.TotalAttempts,
			"completed_attempts": r.CompletedAttempts,
			"total_users":        r.TotalUsers,
			"completion_rate":    r.CompletionRate,
		}
	}

	return map[string]interface{}{
		"assessments": resultMaps,
	}, nil
}

// GetScoreDistribution calculates the distribution of scores across all attempts
func (r *attemptRepository) GetScoreDistribution() (map[string]interface{}, error) {
	type ScoreRange struct {
		Range string `gorm:"column:range"`
		Count int    `gorm:"column:count"`
	}

	var results []ScoreRange

	err := r.db.Raw(`
		WITH score_ranges AS (
			SELECT 
				CASE
					WHEN score >= 90 THEN '90-100'
					WHEN score >= 80 THEN '80-89'
					WHEN score >= 70 THEN '70-79'
					WHEN score >= 60 THEN '60-69'
					WHEN score >= 50 THEN '50-59'
					ELSE 'Below 50'
				END as range,
				COUNT(*) as count
			FROM attempts
			WHERE score IS NOT NULL
			GROUP BY range
		)
		SELECT range, count
		FROM score_ranges
		ORDER BY range
	`).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch score distribution: %w", err)
	}

	// Create an array of distribution for ordered visualization
	var distributionArray []map[string]interface{}
	for _, r := range results {
		distributionArray = append(distributionArray, map[string]interface{}{
			"range": r.Range,
			"count": r.Count,
		})
	}

	return map[string]interface{}{
		"distribution": distributionArray,
	}, nil
}

// GetAverageTimeSpent calculates the average time spent on assessments
func (r *attemptRepository) GetAverageTimeSpent() (map[string]interface{}, error) {
	type TimeSpent struct {
		ID              uint    `gorm:"column:id"`
		Title           string  `gorm:"column:title"`
		AvgMinutes      float64 `gorm:"column:avg_minutes"`
		ExpectedMinutes int     `gorm:"column:expected_minutes"`
	}

	var results []TimeSpent

	err := r.db.Raw(`
		SELECT 
			a.id, a.title,
			AVG(CASE WHEN att.duration IS NOT NULL THEN att.duration ELSE 
				EXTRACT(EPOCH FROM (COALESCE(att.ended_at, att.submitted_at) - att.started_at)) / 60
			END) as avg_minutes,
			a.duration as expected_minutes
		FROM assessments a
		JOIN attempts att ON a.id = att.assessment_id
		WHERE att.status = 'Completed'
		GROUP BY a.id, a.title, a.duration
		ORDER BY avg_minutes DESC
		LIMIT 10
	`).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch average time spent: %w", err)
	}

	// Convert to map[string]interface{}
	resultMaps := make([]map[string]interface{}, len(results))
	for i, r := range results {
		resultMaps[i] = map[string]interface{}{
			"id":               r.ID,
			"title":            r.Title,
			"avg_minutes":      r.AvgMinutes,
			"expected_minutes": r.ExpectedMinutes,
		}
	}

	return map[string]interface{}{
		"timeStats": resultMaps,
	}, nil
}

// GetMostChallengingAssessments finds assessments with the lowest pass rates
func (r *attemptRepository) GetMostChallengingAssessments(limit int) ([]map[string]interface{}, error) {
	type ChallengeStats struct {
		ID             uint    `gorm:"column:id"`
		Title          string  `gorm:"column:title"`
		TotalAttempts  int64   `gorm:"column:total_attempts"`
		PassedAttempts int64   `gorm:"column:passed_attempts"`
		AvgScore       float64 `gorm:"column:avg_score"`
		PassRate       float64 `gorm:"column:pass_rate"`
	}

	var results []ChallengeStats

	err := r.db.Raw(`
		WITH assessment_stats AS (
			SELECT 
				a.id, a.title,
				COUNT(DISTINCT att.id) as total_attempts,
				COUNT(DISTINCT CASE WHEN att.score >= a.passing_score THEN att.id END) as passed_attempts,
				AVG(att.score) as avg_score
			FROM assessments a
			JOIN attempts att ON a.id = att.assessment_id
			WHERE att.status = 'Completed' AND att.score IS NOT NULL
			GROUP BY a.id, a.title, a.passing_score
			HAVING COUNT(att.id) >= 5
		)
		SELECT 
			id, title, total_attempts, passed_attempts, avg_score,
			CASE WHEN total_attempts > 0 THEN 
				ROUND((passed_attempts::numeric / total_attempts::numeric) * 100, 2)
			ELSE 0 END as pass_rate
		FROM assessment_stats
		ORDER BY pass_rate ASC
		LIMIT ?
	`, limit).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch challenging assessments: %w", err)
	}

	// Convert to map[string]interface{}
	resultMaps := make([]map[string]interface{}, len(results))
	for i, r := range results {
		resultMaps[i] = map[string]interface{}{
			"id":              r.ID,
			"title":           r.Title,
			"total_attempts":  r.TotalAttempts,
			"passed_attempts": r.PassedAttempts,
			"avg_score":       r.AvgScore,
			"pass_rate":       r.PassRate,
		}
	}

	return resultMaps, nil
}

// GetMostSuccessfulAssessments finds assessments with the highest pass rates
func (r *attemptRepository) GetMostSuccessfulAssessments(limit int) ([]map[string]interface{}, error) {
	type SuccessStats struct {
		ID             uint    `gorm:"column:id"`
		Title          string  `gorm:"column:title"`
		TotalAttempts  int64   `gorm:"column:total_attempts"`
		PassedAttempts int64   `gorm:"column:passed_attempts"`
		AvgScore       float64 `gorm:"column:avg_score"`
		PassRate       float64 `gorm:"column:pass_rate"`
	}

	var results []SuccessStats

	err := r.db.Raw(`
		WITH assessment_stats AS (
			SELECT 
				a.id, a.title,
				COUNT(DISTINCT att.id) as total_attempts,
				COUNT(DISTINCT CASE WHEN att.score >= a.passing_score THEN att.id END) as passed_attempts,
				AVG(att.score) as avg_score
			FROM assessments a
			JOIN attempts att ON a.id = att.assessment_id
			WHERE att.status = 'Completed' AND att.score IS NOT NULL
			GROUP BY a.id, a.title, a.passing_score
			HAVING COUNT(att.id) >= 5
		)
		SELECT 
			id, title, total_attempts, passed_attempts, avg_score,
			CASE WHEN total_attempts > 0 THEN 
				ROUND((passed_attempts::numeric / total_attempts::numeric) * 100, 2)
			ELSE 0 END as pass_rate
		FROM assessment_stats
		ORDER BY pass_rate DESC
		LIMIT ?
	`, limit).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch successful assessments: %w", err)
	}

	// Convert to map[string]interface{}
	resultMaps := make([]map[string]interface{}, len(results))
	for i, r := range results {
		resultMaps[i] = map[string]interface{}{
			"id":              r.ID,
			"title":           r.Title,
			"total_attempts":  r.TotalAttempts,
			"passed_attempts": r.PassedAttempts,
			"avg_score":       r.AvgScore,
			"pass_rate":       r.PassRate,
		}
	}

	return resultMaps, nil
}

// GetPassRate calculates the overall pass rate for all assessments
func (r *attemptRepository) GetPassRate() (float64, error) {
	type PassRateResult struct {
		PassRate float64 `gorm:"column:pass_rate"`
	}

	var result PassRateResult

	err := r.db.Raw(`
		SELECT
			CASE WHEN COUNT(*) > 0 THEN
				ROUND(
					(COUNT(CASE WHEN att.score >= a.passing_score THEN 1 END)::numeric / COUNT(*)::numeric) * 100,
					2
				)
			ELSE 0 END as pass_rate
		FROM attempts att
		JOIN assessments a ON att.assessment_id = a.id
		WHERE att.status = 'Completed' AND att.score IS NOT NULL
	`).Scan(&result).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate pass rate: %w", err)
	}

	return result.PassRate, nil
}

// CountAll counts all attempts
func (r *attemptRepository) CountAll() (int64, error) {
	var count int64

	result := r.db.Model(&models.Attempt{}).Where("deleted_at IS NULL").Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count attempts: %w", result.Error)
	}

	return count, nil
}

// CountByPeriod counts attempts within the specified number of days
func (r *attemptRepository) CountByPeriod(days int) (int64, error) {
	var count int64

	cutoffDate := time.Now().AddDate(0, 0, -days)

	result := r.db.Model(&models.Attempt{}).
		Where("started_at >= ? AND deleted_at IS NULL", cutoffDate).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to count attempts by period: %w", result.Error)
	}

	return count, nil
}

// SaveSuspiciousActivity saves a suspicious activity record
func (r *attemptRepository) SaveSuspiciousActivity(activity *models.SuspiciousActivity) error {
	now := time.Now()
	activity.CreatedAt = now

	result := r.db.Create(&activity)
	if result.Error != nil {
		return fmt.Errorf("failed to save suspicious activity: %w", result.Error)
	}

	return nil
}

// CountRecentSuspiciousActivity counts suspicious activities within recent hours
func (r *attemptRepository) CountRecentSuspiciousActivity(hours int) (int64, error) {
	var count int64

	cutoffTime := time.Now().Add(time.Duration(-hours) * time.Hour)

	result := r.db.Model(&models.SuspiciousActivity{}).
		Where("timestamp >= ?", cutoffTime).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to count recent suspicious activities: %w", result.Error)
	}

	return count, nil
}

// FindSuspiciousActivitiesByAttemptID finds all suspicious activities for an attempt
func (r *attemptRepository) FindSuspiciousActivitiesByAttemptID(attemptID uint) ([]models.SuspiciousActivity, error) {
	var activities []models.SuspiciousActivity

	result := r.db.Where("attempt_id = ?", attemptID).
		Order("timestamp DESC").
		Find(&activities)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to find suspicious activities: %w", result.Error)
	}

	return activities, nil
}
