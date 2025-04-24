package api

import (
	"assessment_service/configs"
	repository5 "assessment_service/internal/activity/repository"
	service4 "assessment_service/internal/activity/service"
	"assessment_service/internal/assessments/repository/postgres"
	"assessment_service/internal/assessments/service"
	repository4 "assessment_service/internal/attempts/repository"
	service5 "assessment_service/internal/attempts/service"
	"assessment_service/internal/cronjob"
	repository3 "assessment_service/internal/questions/repository"
	service2 "assessment_service/internal/questions/service"
	service3 "assessment_service/internal/student/service"
	"assessment_service/internal/users/repository"
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	router *mux.Router
	config *configs.Config
	log    *zap.Logger
	db     *gorm.DB
}

func NewServer(config *configs.Config, db *gorm.DB, log *zap.Logger) *Server {
	return &Server{
		config: config,
		log:    log,
		db:     db,
	}
}

func (s *Server) Run() error {
	// Set up services and handlers
	// jwtUtil := util.NewJwtImpl()

	// Initialize repositories
	userRepo := repository.NewUserRepository(s.db)
	assessmentRepo := postgres.NewAssessmentRepository(s.db)
	questionRepo := repository3.NewQuestionRepository(s.db)
	attemptRepo := repository4.NewAttemptRepository(s.db)
	activityRepo := repository5.NewActivityRepository(s.db)

	// Initialize services
	assessmentService := service.NewAssessmentService(assessmentRepo, userRepo)
	questionService := service2.NewQuestionService(questionRepo, assessmentRepo)
	studentService := service3.NewStudentService(assessmentRepo, attemptRepo, questionRepo, userRepo, s.log)
	analyticsService := service4.NewAnalyticsService(userRepo, assessmentRepo, attemptRepo, activityRepo, s.log)
	attemptService := service5.NewAttemptService(attemptRepo, s.log)

	// Set up routes
	s.router = SetupRoutes(
		assessmentService,
		questionService,
		analyticsService,
		studentService,
		attemptService,
		s.log,
	)

	port := getEnv("PORT", "8080")

	// Configure server
	srv := &http.Server{
		Addr: ":" + port,
		Handler: handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}))(s.router),
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	// Run server in a goroutine
	go func() {
		s.log.Info(fmt.Sprintf("Server running on port %s", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// setup go-routine for check expired attempts
	cronJob := cron.New(
		cron.WithLogger(
			cron.VerbosePrintfLogger(log.New(os.Stdout, "CRON: ", log.LstdFlags))),
		cron.WithChain(
			cron.Recover(cron.DefaultLogger), // Tự động phục hồi nếu có panic
		))
	cronJobService := cronjob.NewCronJobService(studentService, s.log, cronJob)
	cronJobService.StartAutoSubmit()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	s.log.Info("Shutting down server...")

	// Gracefully shutdown with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		s.log.Fatal("Server shutdown error", zap.Error(err))
		return err
	}

	s.log.Info("Server exited properly")
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
