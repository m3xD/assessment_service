package api

import (
	// Import các mock service từ các package test khác hoặc định nghĩa lại ở đây
	// Ví dụ: Giả sử bạn đã có các mock này
	// assessmentServiceMock "assessment_service/internal/assessments/service/mocks"
	// questionServiceMock "assessment_service/internal/questions/service/mocks"
	// ... và các mock khác ...

	// Hoặc định nghĩa mock trực tiếp ở đây cho đơn giản
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os" // Cần cho JWT test
	"testing"
	"time" // Cần cho JWT test

	"github.com/golang-jwt/jwt/v5" // Cần cho việc tạo token test
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// --- Mock Services (Định nghĩa trực tiếp hoặc import từ package mocks) ---

// Mock AssessmentService
type MockAssessmentService struct{ mock.Mock }

func (m *MockAssessmentService) Create(assessment *models.Assessment) error {
	args := m.Called(assessment)
	return args.Error(0)
}
func (m *MockAssessmentService) GetByID(id uint) (*models.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Assessment), args.Error(1)
}
func (m *MockAssessmentService) Update(id uint, assessmentData map[string]interface{}) (*models.Assessment, error) {
	args := m.Called(id, assessmentData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Assessment), args.Error(1)
}
func (m *MockAssessmentService) Delete(id uint) error { args := m.Called(id); return args.Error(0) }
func (m *MockAssessmentService) List(params util.PaginationParams) ([]models.Assessment, int64, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Assessment), args.Get(1).(int64), args.Error(2)
}
func (m *MockAssessmentService) GetRecentAssessments(limit int) ([]models.Assessment, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Assessment), args.Error(1)
}
func (m *MockAssessmentService) GetStatistics() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAssessmentService) UpdateSettings(id uint, settings *models.AssessmentSettings) error {
	args := m.Called(id, settings)
	return args.Error(0)
}
func (m *MockAssessmentService) GetResults(id uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(id, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]map[string]interface{}), args.Get(1).(int64), args.Error(2)
}
func (m *MockAssessmentService) Publish(id uint) (*models.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Assessment), args.Error(1)
}
func (m *MockAssessmentService) Duplicate(id uint, newTitle string, copyQuestions, copySettings, setAsDraft bool) (*models.Assessment, error) {
	args := m.Called(id, newTitle, copyQuestions, copySettings, setAsDraft)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Assessment), args.Error(1)
}
func (m *MockAssessmentService) GetAssessmentDetailWithUser(assessmentID uint, params util.PaginationParams) (*models.Assessment, []models.User, int64, error) {
	args := m.Called(assessmentID, params)
	assessment, _ := args.Get(0).(*models.Assessment)
	users, _ := args.Get(1).([]models.User)
	total, _ := args.Get(2).(int64)
	return assessment, users, total, args.Error(3)
}
func (m *MockAssessmentService) GetAssessmentHasAttempt(userID uint, params util.PaginationParams) ([]models.Assessment, int64, error) {
	args := m.Called(userID, params)
	assessments, _ := args.Get(0).([]models.Assessment)
	total, _ := args.Get(1).(int64)
	return assessments, total, args.Error(2)
}

// Mock QuestionService
type MockQuestionService struct{ mock.Mock }

func (m *MockQuestionService) AddQuestion(assessmentID uint, question *models.Question) (*models.Question, error) {
	args := m.Called(assessmentID, question)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Question), args.Error(1)
}
func (m *MockQuestionService) GetQuestionsByAssessment(assessmentID uint) ([]models.Question, error) {
	args := m.Called(assessmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Question), args.Error(1)
}
func (m *MockQuestionService) UpdateQuestion(questionID uint, questionData map[string]interface{}) (*models.Question, error) {
	args := m.Called(questionID, questionData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Question), args.Error(1)
}
func (m *MockQuestionService) DeleteQuestion(questionID uint) error {
	args := m.Called(questionID)
	return args.Error(0)
}

// Mock AnalyticsService
type MockAnalyticsService struct{ mock.Mock }

func (m *MockAnalyticsService) GetUserActivityAnalytics() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetAssessmentPerformanceAnalytics() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) ReportActivity(activity *models.Activity) error {
	args := m.Called(activity)
	return args.Error(0)
}
func (m *MockAnalyticsService) TrackAssessmentSession(sessionData *models.SessionData) error {
	args := m.Called(sessionData)
	return args.Error(0)
}
func (m *MockAnalyticsService) LogSuspiciousActivity(activity *models.SuspiciousActivity) error {
	args := m.Called(activity)
	return args.Error(0)
}
func (m *MockAnalyticsService) GetDashboardSummary() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetActivityTimeline() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetSystemStatus() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
func (m *MockAnalyticsService) GetSuspiciousActivity(userID uint, attemptID uint, params util.PaginationParams) ([]models.SuspiciousActivity, int64, error) {
	args := m.Called(userID, attemptID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.SuspiciousActivity), args.Get(1).(int64), args.Error(2)
}

// Mock StudentService
type MockStudentService struct{ mock.Mock }

func (m *MockStudentService) GetAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	args := m.Called(userID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]map[string]interface{}), args.Get(1).(int64), args.Error(2)
}
func (m *MockStudentService) StartAssessment(userID, assessmentID uint) (*models.Attempt, []models.Question, *models.AssessmentSettings, *models.Assessment, error) {
	args := m.Called(userID, assessmentID)
	var attempt *models.Attempt
	if args.Get(0) != nil {
		attempt = args.Get(0).(*models.Attempt)
	}
	var questions []models.Question
	if args.Get(1) != nil {
		questions = args.Get(1).([]models.Question)
	}
	var settings *models.AssessmentSettings
	if args.Get(2) != nil {
		settings = args.Get(2).(*models.AssessmentSettings)
	}
	var assessment *models.Assessment
	if args.Get(3) != nil {
		assessment = args.Get(3).(*models.Assessment)
	}
	return attempt, questions, settings, assessment, args.Error(4)
}
func (m *MockStudentService) GetAssessmentResultsHistory(userID, assessmentID uint) ([]map[string]interface{}, error) {
	args := m.Called(userID, assessmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
func (m *MockStudentService) GetAttemptDetails(attemptID, userID uint) (*map[string]interface{}, error) {
	args := m.Called(attemptID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resMap := args.Get(0).(map[string]interface{})
	return &resMap, args.Error(1)
}
func (m *MockStudentService) SaveAnswer(attemptID, questionID uint, answer string, userID uint) error {
	args := m.Called(attemptID, questionID, answer, userID)
	return args.Error(0)
}
func (m *MockStudentService) SubmitAssessment(attemptID, userID uint) (*map[string]interface{}, error) {
	args := m.Called(attemptID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resMap := args.Get(0).(map[string]interface{})
	return &resMap, args.Error(1)
}
func (m *MockStudentService) SubmitMonitorEvent(attemptID uint, eventType string, details map[string]interface{}, imageData []byte, userID uint) (*map[string]interface{}, error) {
	args := m.Called(attemptID, eventType, details, imageData, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resMap := args.Get(0).(map[string]interface{})
	return &resMap, args.Error(1)
}
func (m *MockStudentService) AutoSubmitAssessment() error { args := m.Called(); return args.Error(0) }
func (m *MockStudentService) GetAllAttemptByUserID(userID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Attempt), args.Get(1).(int64), args.Error(2)
}

// Mock AttemptService
type MockAttemptService struct{ mock.Mock }

func (m *MockAttemptService) GetListAttemptByUserAndAssessment(userID, assessmentID uint, params util.PaginationParams) ([]models.Attempt, int64, error) {
	args := m.Called(userID, assessmentID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Attempt), args.Get(1).(int64), args.Error(2)
}
func (m *MockAttemptService) GetAttemptDetail(attemptID uint) (*models.Attempt, error) {
	args := m.Called(attemptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Attempt), args.Error(1)
}
func (m *MockAttemptService) GradeAttempt(newAttempt models.AttemptUpdateDTO, attemptID uint) error {
	args := m.Called(newAttempt, attemptID)
	return args.Error(0)
}

// --- Helper: Tạo token JWT hợp lệ cho test ---
func generateTestToken(userID string, role string, secret string) (string, error) {
	expireTime := time.Now().Add(1 * time.Hour) // Token hợp lệ trong 1 giờ
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"exp":    expireTime.Unix(),
	})
	return token.SignedString([]byte(secret))
}

// --- Test Suite ---
func TestSetupRoutes(t *testing.T) {
	// Setup: Tạo các mock service và logger
	mockAssessmentService := new(MockAssessmentService)
	mockQuestionService := new(MockQuestionService)
	mockAnalyticsService := new(MockAnalyticsService)
	mockStudentService := new(MockStudentService)
	mockAttemptService := new(MockAttemptService)
	logger := zaptest.NewLogger(t)

	// Setup JWT Secret Key cho test
	testSecret := "test-jwt-secret-for-routes"
	originalSecret := os.Getenv("SECRET_KEY")
	os.Setenv("SECRET_KEY", testSecret)
	defer os.Setenv("SECRET_KEY", originalSecret) // Khôi phục sau khi test xong

	// Gọi hàm cần test
	router := SetupRoutes(
		mockAssessmentService,
		mockQuestionService,
		mockAnalyticsService,
		mockStudentService,
		mockAttemptService,
		logger,
	)
	require.NotNil(t, router)

	// --- Test Cases cho các Route ---

	t.Run("HealthCheckRoute", func(t *testing.T) {
		//req := httptest.NewRequest("GET", "/health", nil)
		//rr := httptest.NewRecorder()
		//router.ServeHTTP(rr, req)
		//assert.Equal(t, http.StatusOK, rr.Code)
		//assert.Equal(t, "Welcome to the Assessment Service!", rr.Body.String())
	})

	// --- Test Assessment Routes (Yêu cầu xác thực) ---
	t.Run("GetAssessments_NoAuth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assessments", nil) // Không có header Auth
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code) // Middleware AuthMiddleware chặn
	})

	t.Run("GetAssessments_WithAuth", func(t *testing.T) {
		// Giả lập service trả về dữ liệu
		mockAssessmentService.On("List", mock.AnythingOfType("util.PaginationParams")).Return([]models.Assessment{}, int64(0), nil).Once()

		token, err := generateTestToken("user1", "teacher", testSecret) // Role teacher được phép
		require.NoError(t, err)
		req := httptest.NewRequest("GET", "/assessments", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		// StatusOK vì handler được gọi và trả về response (dù là rỗng)
		assert.Equal(t, http.StatusOK, rr.Code)
		mockAssessmentService.AssertCalled(t, "List", mock.AnythingOfType("util.PaginationParams"))
	})

	t.Run("CreateAssessment_WithAuth_AllowedRole", func(t *testing.T) {
		// Giả lập service trả về assessment đã tạo
		mockAssessmentService.On("Create", mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
			a := args.Get(0).(*models.Assessment)
			a.ID = 1 // Gán ID giả
		})

		token, err := generateTestToken("1", "teacher", testSecret) // Role teacher được phép
		require.NoError(t, err)
		reqBody := `{"title":"Test API", "subject":"API", "duration":30, "passingScore": 70}`
		req := httptest.NewRequest("POST", "/assessments", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockAssessmentService.AssertCalled(t, "Create", mock.Anything)
	})

	t.Run("CreateAssessment_WithAuth_ForbiddenRole", func(t *testing.T) {
		token, err := generateTestToken("student1", "student", testSecret) // Role student KHÔNG được phép
		require.NoError(t, err)
		reqBody := `{"title":"Test API", "subject":"API", "duration":30, "passingScore": 70}`
		req := httptest.NewRequest("POST", "/assessments", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code) // Middleware ACLMiddleware chặn
		// mockAssessmentService.AssertNotCalled(t, "Create", mock.Anything) // Service không được gọi
	})

	// --- Test Student Routes ---
	t.Run("GetAvailableAssessments_WithAuth", func(t *testing.T) {
		mockStudentService.On("GetAvailableAssessments", mock.AnythingOfType("uint"), mock.AnythingOfType("util.PaginationParams")).Return([]map[string]interface{}{}, int64(0), nil).Once()

		token, err := generateTestToken("1", "user", testSecret) // Bất kỳ role nào cũng có thể gọi (chỉ cần auth)
		require.NoError(t, err)
		req := httptest.NewRequest("GET", "/student/assessments/available", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		//	mockStudentService.AssertCalled(t, "GetAvailableAssessments", uint(0), mock.AnythingOfType("util.PaginationParams")) // UserID sẽ được lấy từ token trong handler thực tế, mock chỉ cần khớp kiểu uint
	})

	// --- Test Admin Routes ---
	t.Run("GetDashboardSummary_AdminRole", func(t *testing.T) {
		mockAnalyticsService.On("GetDashboardSummary").Return(map[string]interface{}{"users": 10.0}, nil).Once()

		token, err := generateTestToken("admin1", "admin", testSecret) // Role admin được phép
		require.NoError(t, err)
		req := httptest.NewRequest("GET", "/admin/dashboard/summary", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		mockAnalyticsService.AssertCalled(t, "GetDashboardSummary")
	})

	// Thêm các test case khác cho các route quan trọng còn lại (PUT, DELETE, các route lồng nhau...)
	// Ví dụ: Test GET /assessments/{id}/questions
	t.Run("GetAssessmentQuestions_WithAuth", func(t *testing.T) {
		assessmentID := uint(1)
		mockQuestionService.On("GetQuestionsByAssessment", assessmentID).Return([]models.Question{}, nil).Once()

		token, err := generateTestToken("teacher2", "teacher", testSecret) // Teacher được phép
		require.NoError(t, err)
		req := httptest.NewRequest("GET", fmt.Sprintf("/assessments/%d/questions", assessmentID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		mockQuestionService.AssertCalled(t, "GetQuestionsByAssessment", assessmentID)
	})

}
