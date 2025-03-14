package main

import (
	"exam_service/api/middleware"
	"exam_service/configs"
	_ "exam_service/docs"
	"exam_service/internal/exam/delivery/rest"
	postgres3 "exam_service/internal/exam/repository/postgres"
	service3 "exam_service/internal/exam/service"
	postgres4 "exam_service/internal/exam_attempt/repository/postgres"
	postgres5 "exam_service/internal/question/repository/postgres"
	"exam_service/internal/util"
	pkg_log "exam_service/pkg/logger"
	pkg_postgres "exam_service/pkg/postgres"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"net/http"
	"os"
)

// @title Exam Service API
// @version 1.0
// @description This service allow user to create exam, start exam, submit exam, and list exam
// @termsOfService http://swagger.io/terms/

// @contact.name Duy Khanh
// @contact.email duykhanh.forwork2108@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

//  @securityDefinitions.apiKey  JWT
//  @in                          header
//  @name                        Authorization
//  @description                 JWT security accessToken. Please add it

// @host exam-service-39fe1c7b96de.herokuapp.com
// @BasePath /api/v1
func main() {
	log := pkg_log.NewLogger().Logger

	err := configs.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	// jwt util
	jwt := util.NewJwtImpl()

	initDB, err := pkg_postgres.PostgresConnect()
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	log.Info("success connect to database")

	// repository layer
	examRepo := postgres3.NewExamRepositoryPostgreSQL(initDB, log)
	examAttempt := postgres4.NewExamAttemptRepositoryPostgres(initDB, log)
	questionRepo := postgres5.NewQuestionRepositoryPostgres(initDB, log)

	// service layer
	examService := service3.NewExamServiceImp(examRepo, examAttempt, questionRepo, log)

	// middleware
	authMiddleware := middleware.NewAuthMiddleware(jwt)
	logMiddleware := middleware.NewLogMiddleware(log)

	// handler
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	// router
	router := mux.NewRouter()
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(os.Getenv("LINK_SYSTEM")+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)
	// router.Use(middleware.CORSMiddleware)
	router.Use(logMiddleware.LoggingMiddleware)

	handler := rest.NewExamHandler(examService, router, authMiddleware, log)
	handler.RegisterRouter()

	log.Info("server started at :" + port)
	err = http.ListenAndServe(":"+port, handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	)(router))
	if err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}
}
