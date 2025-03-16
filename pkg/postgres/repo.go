package database

import (
	"assessment_service/configs"
	models "assessment_service/internal/model"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(config configs.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	// Migrate all models
	return db.AutoMigrate(
		&models.User{},
		&models.Assessment{},
		&models.Question{},
		&models.QuestionOption{},
		&models.Attempt{},
		&models.Answer{},
		&models.Activity{},
		&models.SuspiciousActivity{},
		&models.AssessmentSettings{},
	)
}

// Close the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
