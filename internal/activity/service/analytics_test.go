package service

import (
	// Không import mock repo nữa
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require" // Dùng require khi cần
	"go.uber.org/zap/zaptest"
	"testing"
)

// --- Mock UserRepository ---
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockUserRepository) List(params util.PaginationParams) ([]models.User, int64, error) {
	args := m.Called(params)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}
func (m *MockUserRepository) UpdateLastLogin(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockUserRepository) GetUserStats() (int64, int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}
func (m *MockUserRepository) CountAll() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockUserRepository) GetNewUsersCount(days int) (int64, error) {
	args := m.Called(days)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockUserRepository) GetListUserByAssessment(params util.PaginationParams, assessmentID uint) ([]models.User, int64, error) {
	args := m.Called(params, assessmentID)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

// --- Mock AssessmentRepository ---
type MockAssessmentRepository struct {
	mock.Mock
}

func (m *MockAssessmentRepository) Create(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}
func (m *MockAssessmentRepository) FindByID(id uint) (*models.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Assessment), args.Error(1)
}
func (m *MockAssessmentRepository) Update(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}
func (m *MockAssessmentRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockAssessmentRepository) List(params util.PaginationParams) ([]models.Assessment, int64, error) {
	args := m.Called(params)
	return args.Get(0).([]models.Assessment), args.Get(1).(int64), args.Error(2)
}
func (m *MockAssessmentRepository) FindRecent(limit int) ([]models.Assessment, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Assessment), args.Error(1)
}
func (m *MockAssessmentRepository) GetStatistics() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAssessmentRepository) UpdateSettings(id uint, settings *models.AssessmentSettings) error {
	args := m.Called(id, settings)
	return args.Error(0)
}
func (m *MockAssessmentRepository) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(id, params)
	return args.Get(0).([]map[string]interface{}), args.Get(1).(int64), args.Error(2)
}
func (m *MockAssessmentRepository) Publish(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockAssessmentRepository) Duplicate(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}
func (m *MockAssessmentRepository) GetAssessmentHasAttemptByUser(params util.PaginationParams, userID uint) ([]models.Assessment, int64, error) {
	args := m.Called(params, userID)
	return args.Get(0).([]models.Assessment), args.Get(1).(int64), args.Error(2)
}

// --- Mock AttemptRepository ---
type MockAttemptRepository struct {
	mock.Mock
}

// Implement các phương thức của AttemptRepository interface
func (m *MockAttemptRepository) Create(attempt *models.Attempt) error {
	args := m.Called(attempt)
	return args.Error(0)
}
func (m *MockAttemptRepository) FindByID(id uint) (*models.Attempt, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Attempt), args.Error(1)
}
func (m *MockAttemptRepository) Update(attempt *models.Attempt) error {
	args := m.Called(attempt)
	return args.Error(0)
}
func (m *MockAttemptRepository) Delete(id uint) error { args := m.Called(id); return args.Error(0) }
func (m *MockAttemptRepository) SaveAnswer(answer *models.Answer) error {
	args := m.Called(answer)
	return args.Error(0)
}
func (m *MockAttemptRepository) UpdateAnswer(answer *models.Answer) error {
	args := m.Called(answer)
	return args.Error(0)
}
func (m *MockAttemptRepository) FindAnswersByAttemptID(attemptID uint) ([]models.Answer, error) {
	args := m.Called(attemptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Answer), args.Error(1)
}
func (m *MockAttemptRepository) FindAnswerByAttemptAndQuestion(attemptID, questionID uint) (*models.Answer, error) {
	args := m.Called(attemptID, questionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Answer), args.Error(1)
}
func (m *MockAttemptRepository) FindAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(userID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]map[string]interface{}), args.Get(1).(int64), args.Error(2)
}
func (m *MockAttemptRepository) HasCompletedAssessment(userID, assessmentID uint) (bool, error) {
	args := m.Called(userID, assessmentID)
	return args.Bool(0), args.Error(1)
}
func (m *MockAttemptRepository) CountAttemptsByUserAndAssessment(userID, assessmentID uint) (int, error) {
	args := m.Called(userID, assessmentID)
	return args.Int(0), args.Error(1)
}
func (m *MockAttemptRepository) FindCompletedAttemptsByUserAndAssessment(userID, assessmentID uint) ([]map[string]interface{}, error) {
	args := m.Called(userID, assessmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockAttemptRepository) GetAllAttemptByUserId(userID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Attempt), args.Get(1).(int64), args.Error(2)
}
func (m *MockAttemptRepository) ListAttemptByUserAndAssessmentID(userID uint, assessmentID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, assessmentID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Attempt), args.Get(1).(int64), args.Error(2)
}
func (m *MockAttemptRepository) GetAssessmentCompletionRates() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAttemptRepository) GetScoreDistribution() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAttemptRepository) GetAverageTimeSpent() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAttemptRepository) GetMostChallengingAssessments(limit int) ([]map[string]interface{}, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockAttemptRepository) GetMostSuccessfulAssessments(limit int) ([]map[string]interface{}, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockAttemptRepository) GetPassRate() (float64, error) {
	args := m.Called()
	val, _ := args.Get(0).(float64)
	return val, args.Error(1)
}
func (m *MockAttemptRepository) CountAll() (int64, error) {
	args := m.Called()
	val, _ := args.Get(0).(int64)
	return val, args.Error(1)
}
func (m *MockAttemptRepository) CountByPeriod(days int) (int64, error) {
	args := m.Called(days)
	val, _ := args.Get(0).(int64)
	return val, args.Error(1)
}
func (m *MockAttemptRepository) SaveSuspiciousActivity(activity *models.SuspiciousActivity) error {
	args := m.Called(activity)
	return args.Error(0)
}
func (m *MockAttemptRepository) CountRecentSuspiciousActivity(hours int) (int64, error) {
	args := m.Called(hours)
	val, _ := args.Get(0).(int64)
	return val, args.Error(1)
}
func (m *MockAttemptRepository) FindSuspiciousActivitiesByAttemptID(attemptID uint) ([]models.SuspiciousActivity, error) {
	args := m.Called(attemptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SuspiciousActivity), args.Error(1)
}
func (m *MockAttemptRepository) ExpiredAttempt() ([]models.Attempt, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Attempt), args.Error(1)
}
func (m *MockAttemptRepository) IsUserInAttempt(userID uint) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

// --- Mock ActivityRepository ---
type MockActivityRepository struct {
	mock.Mock
}

func (m *MockActivityRepository) Create(activity *models.Activity) error {
	args := m.Called(activity)
	return args.Error(0)
}
func (m *MockActivityRepository) FindByUserID(userID uint, params util.PaginationParams) ([]models.Activity, int64, error) {
	args := m.Called(userID, params)
	return args.Get(0).([]models.Activity), args.Get(1).(int64), args.Error(2)
}
func (m *MockActivityRepository) GetDailyActiveUsers(days int) ([]map[string]interface{}, error) {
	args := m.Called(days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockActivityRepository) GetActivityByHour() ([]map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockActivityRepository) GetActivityByType() ([]map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockActivityRepository) GetTotalActiveUsers() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockActivityRepository) GetRecentActivity(hours int) ([]map[string]interface{}, error) {
	args := m.Called(hours)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockActivityRepository) GetActiveUsers(minutes int) (int64, error) {
	args := m.Called(minutes)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockActivityRepository) BulkCreate(activities []models.Activity) error {
	args := m.Called(activities)
	return args.Error(0)
}
func (m *MockActivityRepository) FindByAssessmentID(assessmentID uint, params util.PaginationParams) ([]models.Activity, int64, error) {
	args := m.Called(assessmentID, params)
	return args.Get(0).([]models.Activity), args.Get(1).(int64), args.Error(2)
}
func (m *MockActivityRepository) FindSuspiciousActivity(userID uint, attemptID uint, params util.PaginationParams) ([]models.SuspiciousActivity, int64, error) {
	args := m.Called(userID, attemptID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.SuspiciousActivity), args.Get(1).(int64), args.Error(2)
}
func (m *MockActivityRepository) CountByPeriod(days int) (int64, error) {
	args := m.Called(days)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockActivityRepository) GetTrending() ([]map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// --- Test Cases ---

func TestAnalyticsService_GetUserActivityAnalytics(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	// Cung cấp các mock repo cần thiết
	service := NewAnalyticsService(mockUserRepo, nil, nil, mockActivityRepo, logger)

	expectedDailyActive := []map[string]interface{}{{"date": "2023-01-01", "count": 10.0}}
	expectedActivityByHour := []map[string]interface{}{{"hour": 10.0, "count": 5.0}}
	expectedActivityByType := []map[string]interface{}{{"type": "LOGIN", "count": 20.0}}
	expectedTotalActive := int64(50)
	expectedNewUsers := int64(5)

	mockActivityRepo.On("GetDailyActiveUsers", 7).Return(expectedDailyActive, nil)
	mockActivityRepo.On("GetActivityByHour").Return(expectedActivityByHour, nil)
	mockActivityRepo.On("GetActivityByType").Return(expectedActivityByType, nil)
	mockActivityRepo.On("GetTotalActiveUsers").Return(expectedTotalActive, nil)
	mockUserRepo.On("GetNewUsersCount", 7).Return(expectedNewUsers, nil)

	analytics, err := service.GetUserActivityAnalytics()

	assert.NoError(t, err)
	require.NotNil(t, analytics)
	assert.Equal(t, expectedDailyActive, analytics["dailyActiveUsers"])
	assert.Equal(t, expectedActivityByHour, analytics["activityByHour"])
	assert.Equal(t, expectedActivityByType, analytics["activityByType"])
	assert.Equal(t, expectedTotalActive, analytics["totalActiveUsers"])
	assert.Equal(t, expectedNewUsers, analytics["newUsersLastWeek"])

	mockActivityRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetAssessmentPerformanceAnalytics(t *testing.T) {
	mockAttemptRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	// Cung cấp mock repo cần thiết
	service := NewAnalyticsService(nil, nil, mockAttemptRepo, nil, logger)

	expectedCompletionRates := map[string]interface{}{"rate": 75.0}
	expectedScoreDist := map[string]interface{}{"dist": []int{1, 2, 3}}
	expectedAvgTime := map[string]interface{}{"avg": 30.5}
	expectedChallenging := []map[string]interface{}{{"id": 1, "title": "Hard one"}}
	expectedSuccessful := []map[string]interface{}{{"id": 2, "title": "Easy one"}}

	mockAttemptRepo.On("GetAssessmentCompletionRates").Return(expectedCompletionRates, nil)
	mockAttemptRepo.On("GetScoreDistribution").Return(expectedScoreDist, nil)
	mockAttemptRepo.On("GetAverageTimeSpent").Return(expectedAvgTime, nil)
	mockAttemptRepo.On("GetMostChallengingAssessments", 2).Return(expectedChallenging, nil)
	mockAttemptRepo.On("GetMostSuccessfulAssessments", 2).Return(expectedSuccessful, nil)

	analytics, err := service.GetAssessmentPerformanceAnalytics()
	assert.NoError(t, err)
	require.NotNil(t, analytics)
	assert.Equal(t, expectedCompletionRates, analytics["assessmentCompletionRates"])
	assert.Equal(t, expectedScoreDist, analytics["scoreDistribution"])
	assert.Equal(t, expectedAvgTime, analytics["averageTimeSpent"])
	assert.Equal(t, expectedChallenging, analytics["mostChallenging"])
	assert.Equal(t, expectedSuccessful, analytics["mostSuccessful"])

	mockAttemptRepo.AssertExpectations(t)
}

func TestAnalyticsService_ReportActivity(t *testing.T) {
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, nil, nil, mockActivityRepo, logger)

	activity := &models.Activity{UserID: 1, Action: "TEST_ACTION"}

	mockActivityRepo.On("Create", activity).Return(nil)

	err := service.ReportActivity(activity)
	assert.NoError(t, err)
	mockActivityRepo.AssertExpectations(t)
	assert.NotZero(t, activity.Timestamp) // Timestamp should be set if it was zero
}

func TestAnalyticsService_TrackAssessmentSession(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, mockAssessmentRepo, nil, mockActivityRepo, logger)

	assessmentID := uint(1)
	sessionData := &models.SessionData{UserID: 1, AssessmentID: assessmentID, Action: "SESSION_START"}
	expectedActivity := &models.Activity{
		UserID:       sessionData.UserID,
		Action:       sessionData.Action,
		AssessmentID: &sessionData.AssessmentID,
		Details:      sessionData.Details,
		UserAgent:    sessionData.UserAgent,
		// Timestamp sẽ được set trong service
	}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(&models.Assessment{ID: assessmentID}, nil)
	mockActivityRepo.On("Create", mock.MatchedBy(func(act *models.Activity) bool {
		return act.UserID == expectedActivity.UserID &&
			act.Action == expectedActivity.Action &&
			act.AssessmentID != nil && *act.AssessmentID == *expectedActivity.AssessmentID
	})).Return(nil)

	err := service.TrackAssessmentSession(sessionData)
	assert.NoError(t, err)
	mockAssessmentRepo.AssertExpectations(t)
	mockActivityRepo.AssertExpectations(t)
}

func TestAnalyticsService_TrackAssessmentSession_AssessmentNotFound(t *testing.T) {
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockActivityRepo := new(MockActivityRepository) // Cần mock này dù không gọi hàm nào của nó
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, mockAssessmentRepo, nil, mockActivityRepo, logger)

	assessmentID := uint(99)
	sessionData := &models.SessionData{UserID: 1, AssessmentID: assessmentID, Action: "SESSION_START"}

	mockAssessmentRepo.On("FindByID", assessmentID).Return(nil, errors.New("not found"))

	err := service.TrackAssessmentSession(sessionData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockAssessmentRepo.AssertExpectations(t)
	mockActivityRepo.AssertNotCalled(t, "Create", mock.Anything) // Đảm bảo Create không được gọi
}

func TestAnalyticsService_LogSuspiciousActivity(t *testing.T) {
	mockAttemptRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, nil, mockAttemptRepo, nil, logger)

	activity := &models.SuspiciousActivity{UserID: 1, Type: "TAB_SWITCH"}

	mockAttemptRepo.On("SaveSuspiciousActivity", activity).Return(nil)

	err := service.LogSuspiciousActivity(activity)
	assert.NoError(t, err)
	mockAttemptRepo.AssertExpectations(t)
	assert.NotZero(t, activity.Timestamp) // Timestamp should be set
}

func TestAnalyticsService_GetDashboardSummary(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(mockUserRepo, mockAssessmentRepo, mockAttemptRepo, mockActivityRepo, logger)

	mockUserRepo.On("CountAll").Return(int64(100), nil)
	mockUserRepo.On("GetUserStats").Return(int64(80), int64(20), nil) // active, inactive
	mockUserRepo.On("GetNewUsersCount", 7).Return(int64(10), nil)
	mockAssessmentRepo.On("GetStatistics").Return(map[string]interface{}{
		"totalAssessments":   int64(50),
		"activeAssessments":  int64(30),
		"draftAssessments":   int64(15),
		"expiredAssessments": int64(5),
	}, nil)
	mockAttemptRepo.On("CountAll").Return(int64(500), nil)
	mockAttemptRepo.On("CountByPeriod", 7).Return(int64(50), nil)
	mockAttemptRepo.On("GetPassRate").Return(75.0, nil)
	mockActivityRepo.On("GetActiveUsers", 15).Return(int64(12), nil)              // users online in last 15 mins
	mockAttemptRepo.On("CountRecentSuspiciousActivity", 24).Return(int64(3), nil) // suspicious in last 24h

	summary, err := service.GetDashboardSummary()
	assert.NoError(t, err)
	require.NotNil(t, summary)
	// Add more specific assertions for the structure of the summary map
	usersSummary, ok := summary["users"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, int64(100), usersSummary["total"])
	// ... thêm các assertion khác cho assessments, activity ...

	mockUserRepo.AssertExpectations(t)
	mockAssessmentRepo.AssertExpectations(t)
	mockAttemptRepo.AssertExpectations(t)
	mockActivityRepo.AssertExpectations(t)
}

// Test các trường hợp lỗi cho GetDashboardSummary
func TestAnalyticsService_GetDashboardSummary_UserRepoError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	// ... các mock khác ...
	mockActivityRepo := new(MockActivityRepository)
	mockAssessmentRepo := new(MockAssessmentRepository)
	mockAttemptRepo := new(MockAttemptRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(mockUserRepo, mockAssessmentRepo, mockAttemptRepo, mockActivityRepo, logger)

	mockUserRepo.On("CountAll").Return(int64(0), errors.New("user count error")) // Giả lập lỗi ở đây
	// Các expectation khác có thể không cần nếu lỗi xảy ra sớm

	summary, err := service.GetDashboardSummary()
	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "user count error")
	mockUserRepo.AssertExpectations(t)
	// Các repo khác có thể chưa được gọi
}

// ... (Thêm các test case lỗi tương tự cho AssessmentRepo, AttemptRepo, ActivityRepo) ...

func TestAnalyticsService_GetActivityTimeline(t *testing.T) {
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, nil, nil, mockActivityRepo, logger)

	expectedTimeline := []map[string]interface{}{{"event": "LOGIN"}}
	mockActivityRepo.On("GetRecentActivity", 48).Return(expectedTimeline, nil)

	timeline, err := service.GetActivityTimeline()
	assert.NoError(t, err)
	require.NotNil(t, timeline)
	assert.Equal(t, expectedTimeline, timeline["timeline"])
	mockActivityRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetActivityTimeline_Error(t *testing.T) {
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, nil, nil, mockActivityRepo, logger)

	mockActivityRepo.On("GetRecentActivity", 48).Return(nil, errors.New("timeline error"))

	timeline, err := service.GetActivityTimeline()
	assert.Error(t, err)
	assert.Nil(t, timeline)
	assert.Contains(t, err.Error(), "timeline error")
	mockActivityRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetSystemStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, nil, nil, nil, logger) // No repo calls for this one

	status, err := service.GetSystemStatus()
	assert.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "healthy", status["status"])
	services, ok := status["services"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "operational", services["database"])
}

func TestAnalyticsService_GetSuspiciousActivity(t *testing.T) {
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, nil, nil, mockActivityRepo, logger)

	userID := uint(1)
	attemptID := uint(10)
	params := util.PaginationParams{Page: 0, Limit: 10}
	expectedActivities := []models.SuspiciousActivity{{ID: 1, UserID: userID, AttemptID: attemptID}}
	expectedTotal := int64(1)

	mockActivityRepo.On("FindSuspiciousActivity", userID, attemptID, params).Return(expectedActivities, expectedTotal, nil)

	activities, total, err := service.GetSuspiciousActivity(userID, attemptID, params)
	assert.NoError(t, err)
	assert.Equal(t, expectedActivities, activities)
	assert.Equal(t, expectedTotal, total)
	mockActivityRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetSuspiciousActivity_Error(t *testing.T) {
	mockActivityRepo := new(MockActivityRepository)
	logger := zaptest.NewLogger(t)
	service := NewAnalyticsService(nil, nil, nil, mockActivityRepo, logger)

	userID := uint(1)
	attemptID := uint(10)
	params := util.PaginationParams{Page: 0, Limit: 10}
	repoError := errors.New("find suspicious error")

	mockActivityRepo.On("FindSuspiciousActivity", userID, attemptID, params).Return(nil, int64(0), repoError)

	activities, total, err := service.GetSuspiciousActivity(userID, attemptID, params)
	assert.Error(t, err)
	assert.Nil(t, activities)
	assert.Zero(t, total)
	assert.Equal(t, repoError, err)
	mockActivityRepo.AssertExpectations(t)
}
