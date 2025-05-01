package api

import (
	"assessment_service/internal/activity/delivery/rest"
	"assessment_service/internal/activity/service"
	assessment_handler "assessment_service/internal/assessments/delivery/rest"
	assessment_service "assessment_service/internal/assessments/service"
	"assessment_service/internal/attempts/delivery"
	service3 "assessment_service/internal/attempts/service"
	"assessment_service/internal/middleware"
	question_handler "assessment_service/internal/questions/delivery/rest"
	question_service "assessment_service/internal/questions/service"
	rest2 "assessment_service/internal/student/delivery/rest"
	service2 "assessment_service/internal/student/service"
	"assessment_service/internal/util"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

func SetupRoutes(
	assessmentService assessment_service.AssessmentService,
	questionService question_service.QuestionService,
	analyticsService service.AnalyticsService,
	studentService service2.StudentService,
	attemptService service3.AttemptService,
	log *zap.Logger,
) *mux.Router {
	router := mux.NewRouter()
	jwtService := util.NewJwtImpl()

	loggingMiddleware := middleware.NewLogMiddleware(log)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	router.Use(loggingMiddleware.LoggingMiddleware)
	router.Use(authMiddleware.AuthMiddleware())
	router.Use(middleware.CORSMiddleware)

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to the Assessment Service!"))
	})
	// router.Use(authMiddleware.OwnerMiddleware())
	// Assessments
	assessmentHandler := assessment_handler.NewAssessmentHandler(assessmentService, log)
	questionHandler := question_handler.NewQuestionHandler(questionService, log)
	analyticsHandler := rest.NewAnalyticsHandler(analyticsService)
	studentHandler := rest2.NewStudentHandler(studentService, log)
	attemptHandler := delivery.NewAttemptHandler(attemptService, log)

	assessmentsRouter := router.PathPrefix("/assessments").Subrouter()
	{
		// General assessment routes
		assessmentsRouter.HandleFunc("", assessmentHandler.ListAssessments).Methods("GET")
		assessmentsRouter.HandleFunc("", assessmentHandler.CreateAssessment).Methods("POST")
		assessmentsRouter.Use(authMiddleware.ACLMiddleware("admin", "teacher"))

		assessmentsRouter.HandleFunc("/{id:[0-9]+}", assessmentHandler.GetAssessmentById).Methods("GET")
		assessmentsRouter.HandleFunc("/{id:[0-9]+}", assessmentHandler.UpdateAssessment).Methods("PUT")
		assessmentsRouter.HandleFunc("/{id:[0-9]+}", assessmentHandler.DeleteAssessment).Methods("DELETE")
		assessmentsRouter.HandleFunc("/{id:[0-9]+}/duplicate", assessmentHandler.DuplicateAssessment).Methods("POST")
		assessmentsRouter.HandleFunc("/{id:[0-9]+}/settings", assessmentHandler.UpdateSettings).Methods("PUT")
		assessmentsRouter.HandleFunc("/{id:[0-9]+}/results", assessmentHandler.GetAssessmentResults).Methods("GET")
		assessmentsRouter.HandleFunc("/{id:[0-9]+}/publish", assessmentHandler.PublishAssessment).Methods("POST")

		// Question routes (nested under assessments)
		assessmentsRouter.HandleFunc("/{id:[0-9]+}/questions", questionHandler.GetQuestionsByAssessment).Methods("GET")
		assessmentsRouter.HandleFunc("/{id:[0-9]+}/questions", questionHandler.AddQuestion).Methods("POST")
		assessmentsRouter.HandleFunc("/{assessmentId:[0-9]+}/questions/{questionId:[0-9]+}", questionHandler.UpdateQuestion).Methods("PUT")
		assessmentsRouter.HandleFunc("/{assessmentId:[0-9]+}/questions/{questionId:[0-9]+}", questionHandler.DeleteQuestion).Methods("DELETE")

		// Statistics and recent assessments
		assessmentsRouter.HandleFunc("/recent", assessmentHandler.GetRecentAssessments).Methods("GET")
		assessmentsRouter.HandleFunc("/statistics", assessmentHandler.GetAssessmentStatistics).Methods("GET")
	}

	// Analytics
	analyticsRouter := router.PathPrefix("/analytics").Subrouter()

	// Analytics routes for teachers and admins
	analyticsTeacherRouter := analyticsRouter.PathPrefix("").Subrouter()
	analyticsTeacherRouter.Use(authMiddleware.ACLMiddleware("admin", "teacher"))
	analyticsTeacherRouter.HandleFunc("/user-activity", analyticsHandler.GetUserActivityAnalytics).Methods("GET")
	analyticsTeacherRouter.HandleFunc("/assessment-performance", analyticsHandler.GetAssessmentPerformanceAnalytics).Methods("GET")

	// General analytics routes (for all authenticated users)
	analyticsRouter.HandleFunc("/activity", analyticsHandler.ReportActivity).Methods("POST")
	analyticsRouter.HandleFunc("/assessments/{id:[0-9]+}/session", analyticsHandler.TrackAssessmentSession).Methods("POST")
	analyticsRouter.HandleFunc("/suspicious", analyticsHandler.LogSuspiciousActivity).Methods("POST")

	// Admin dashboard routes
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware.ACLMiddleware("admin"))
	adminRouter.HandleFunc("/dashboard/summary", analyticsHandler.GetDashboardSummary).Methods("GET")
	adminRouter.HandleFunc("/dashboard/activity", analyticsHandler.GetActivityTimeline).Methods("GET")
	adminRouter.HandleFunc("/system/status", analyticsHandler.GetSystemStatus).Methods("GET")
	adminRouter.HandleFunc("/attempt/grade/{attemptID:[0-9]+}", attemptHandler.GradeAttempt).Methods("POST")
	adminRouter.HandleFunc("/attempts/{assessmentID:[0-9]+}/users/{userID:[0-9]+}", attemptHandler.GetListAttemptByUserAndAssessment).Methods("GET")
	adminRouter.HandleFunc("/attempt/{attemptID:[0-9]+}/users/{userID:[0-9]+}", attemptHandler.GetAttemptDetail).Methods("GET")
	adminRouter.HandleFunc("/users/{userID:[0-9]+}/attempts", studentHandler.GetAllAttemptForUser).Methods("GET")
	adminRouter.HandleFunc("/assessments/{assessmentID:[0-9]+}", assessmentHandler.GetAssessmentWithUserHasAttempt).Methods("GET")
	adminRouter.HandleFunc("/activity/{userID:[0-9]+}", analyticsHandler.GetSuspiciousActivity).Methods("GET")
	adminRouter.HandleFunc("/assessments/attempted/{userID:[0-9]+}", assessmentHandler.GetAssessmentHasBeenAttemptByUser).Methods("GET")

	// Student routes (for taking assessments)
	studentRouter := router.PathPrefix("/student").Subrouter()
	studentRouter.HandleFunc("/assessments/available", studentHandler.GetAvailableAssessments).Methods("GET")
	studentRouter.HandleFunc("/assessments/{id:[0-9]+}/start", studentHandler.StartAssessment).Methods("POST")
	studentRouter.HandleFunc("/assessments/{id:[0-9]+}/results", studentHandler.GetAssessmentResultsHistory).Methods("GET")
	studentRouter.HandleFunc("/attempts/{attemptId:[0-9]+}", studentHandler.GetAttemptDetails).Methods("GET")
	studentRouter.HandleFunc("/attempts/{attemptId:[0-9]+}/answers", studentHandler.SaveAnswer).Methods("POST")
	studentRouter.HandleFunc("/attempts/{attemptId:[0-9]+}/submit", studentHandler.SubmitAssessment).Methods("POST")
	studentRouter.HandleFunc("/attempts/{attemptId:[0-9]+}/monitor", studentHandler.SubmitMonitorEvent).Methods("POST")

	return router
}
