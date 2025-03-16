package main

import (
	"assessment_service/configs"
	"assessment_service/internal/api"
	pkg "assessment_service/pkg/logger"
	database "assessment_service/pkg/postgres"
	"go.uber.org/zap"
	"log"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	// Initialize the logger
	log := pkg.NewLogger().Logger

	// Initialize the database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("failed to connect to database:", zap.Error(err))
	}

	// Run database migrations
	/*err = database.RunMigrations(db)
	if err != nil {
		log.Fatal("failed to run migrations:", zap.Error(err))
	}*/

	server := api.NewServer(cfg, db, log)

	// Start the server
	err = server.Run()
	if err != nil {
		log.Fatal("failed to start server:", zap.Error(err))
	}
}
