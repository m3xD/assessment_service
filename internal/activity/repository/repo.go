package repository

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"fmt"

	"time"

	"gorm.io/gorm"
)

type ActivityRepository interface {
	Create(activity *models.Activity) error
	FindByUserID(userID uint, params util.PaginationParams) ([]models.Activity, int64, error)
	GetDailyActiveUsers(days int) ([]map[string]interface{}, error)
	GetActivityByHour() ([]map[string]interface{}, error)
	GetActivityByType() ([]map[string]interface{}, error)
	GetTotalActiveUsers() (int64, error)
	GetRecentActivity(hours int) ([]map[string]interface{}, error)
	GetActiveUsers(minutes int) (int64, error)
	BulkCreate(activities []models.Activity) error
	FindByAssessmentID(assessmentID uint, params util.PaginationParams) ([]models.Activity, int64, error)
	CountByPeriod(days int) (int64, error)
	GetTrending() ([]map[string]interface{}, error)
}

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) Create(activity *models.Activity) error {
	return r.db.Create(activity).Error
}

func (r *activityRepository) FindByUserID(userID uint, params util.PaginationParams) ([]models.Activity, int64, error) {
	var activities []models.Activity
	var total int64

	query := r.db.Model(&models.Activity{}).Where("user_id = ?", userID)

	// Apply date range filter if provided
	if params.Filters != nil {
		if from, ok := params.Filters["from"].(string); ok && from != "" {
			query = query.Where("timestamp >= ?", from)
		}

		if to, ok := params.Filters["to"].(string); ok && to != "" {
			query = query.Where("timestamp <= ?", to)
		}
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortDir)
	} else {
		query = query.Order("timestamp DESC")
	}

	query = query.Offset(params.Offset).Limit(params.Limit)

	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

func (r *activityRepository) GetDailyActiveUsers(days int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// Calculate start date
	startDate := time.Now().UTC().AddDate(0, 0, -days)

	// SQL query to get active users count by day
	query := `
		SELECT 
			DATE(timestamp) as date, 
			COUNT(DISTINCT user_id) as count 
		FROM 
			activities 
		WHERE 
			timestamp >= ? 
		GROUP BY 
			DATE(timestamp) 
		ORDER BY 
			date
	`

	err := r.db.Raw(query, startDate).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *activityRepository) GetActivityByHour() ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// SQL query to get activity count by hour of day
	query := `
		SELECT 
			EXTRACT(HOUR FROM timestamp) as hour, 
			COUNT(*) as count 
		FROM 
			activities 
		WHERE 
			timestamp >= NOW() - INTERVAL '7 days' 
		GROUP BY 
			EXTRACT(HOUR FROM timestamp) 
		ORDER BY 
			hour
	`

	err := r.db.Raw(query).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *activityRepository) GetActivityByType() ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// SQL query to get activity count by action type
	query := `
		SELECT 
			action as type, 
			COUNT(*) as count 
		FROM 
			activities 
		WHERE 
			timestamp >= NOW() - INTERVAL '30 days' 
		GROUP BY 
			action 
		ORDER BY 
			count DESC
	`

	err := r.db.Raw(query).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *activityRepository) GetTotalActiveUsers() (int64, error) {
	var count int64

	// Get count of unique users that had activity in last 30 days
	err := r.db.Model(&models.Activity{}).
		Where("timestamp >= NOW() - INTERVAL '30 days'").
		Distinct("user_id").
		Count(&count).Error

	return count, err
}

func (r *activityRepository) GetRecentActivity(hours int) ([]map[string]interface{}, error) {
	var activities []models.Activity
	var result []map[string]interface{}

	// Get recent activities
	err := r.db.
		Preload("User").
		Where("timestamp >= ?", time.Now().UTC().Add(-time.Hour*time.Duration(hours))).
		Order("timestamp DESC").
		Limit(50).
		Find(&activities).Error

	if err != nil {
		return nil, err
	}

	// Transform to response format
	for _, activity := range activities {
		var eventType string

		switch activity.Action {
		case "LOGIN":
			eventType = "USER_ACTIVITY"
		case "ASSESSMENT_START":
			eventType = "ASSESSMENT_STARTED"
		case "ASSESSMENT_SUBMIT":
			eventType = "ASSESSMENT_COMPLETED"
		case "SUSPICIOUS_ACTIVITY":
			eventType = "SUSPICIOUS_ACTIVITY"
		default:
			eventType = "OTHER"
		}

		item := map[string]interface{}{
			"id":        activity.ID,
			"type":      eventType,
			"user":      activity.User.Name,
			"userId":    activity.UserID,
			"details":   activity.Details,
			"timestamp": activity.Timestamp,
		}

		if activity.AssessmentID != nil {
			// In a real implementation, we would join with Assessment table
			// to get the assessment title
			item["assessmentId"] = *activity.AssessmentID
			item["assessment"] = "Assessment " + fmt.Sprint(*activity.AssessmentID)
		}

		result = append(result, item)
	}

	return result, nil
}

func (r *activityRepository) GetActiveUsers(minutes int) (int64, error) {
	var count int64

	// Get count of unique users active in the past X minutes
	err := r.db.Model(&models.Activity{}).
		Where("timestamp >= NOW() - INTERVAL '? minutes'", minutes).
		Distinct("user_id").
		Count(&count).Error

	return count, err
}

func (r *activityRepository) BulkCreate(activities []models.Activity) error {
	return r.db.CreateInBatches(activities, 100).Error
}

func (r *activityRepository) FindByAssessmentID(assessmentID uint, params util.PaginationParams) ([]models.Activity, int64, error) {
	var activities []models.Activity
	var total int64

	query := r.db.Model(&models.Activity{}).Where("assessment_id = ?", assessmentID)

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortDir)
	} else {
		query = query.Order("timestamp DESC")
	}

	query = query.Offset(params.Offset).Limit(params.Limit)

	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

func (r *activityRepository) CountByPeriod(days int) (int64, error) {
	var count int64

	err := r.db.Model(&models.Activity{}).
		Where("timestamp >= NOW() - INTERVAL '? days'", days).
		Count(&count).Error

	return count, err
}

func (r *activityRepository) GetTrending() ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	// SQL query to get trending assessments (most activity in last 7 days)
	query := `
		SELECT 
			a.assessment_id as id, 
			ass.title as title,
			COUNT(*) as activity_count 
		FROM 
			activities a
		JOIN
			assessments ass ON a.assessment_id = ass.id
		WHERE 
			a.assessment_id IS NOT NULL AND
			a.timestamp >= NOW() - INTERVAL '7 days' 
		GROUP BY 
			a.assessment_id, ass.title
		ORDER BY 
			activity_count DESC
		LIMIT 5
	`

	err := r.db.Raw(query).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
