package cronjob

import (
	"assessment_service/internal/student/service"
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CronJobService struct {
	student service.StudentService
	log     *zap.Logger
	cron    *cron.Cron
}

func NewCronJobService(student service.StudentService, log *zap.Logger, cron *cron.Cron) *CronJobService {
	return &CronJobService{student: student, log: log, cron: cron}
}

func (c *CronJobService) StartAutoSubmit() {
	job, err := c.cron.AddJob("*/5 * * * *", cron.FuncJob(func() {
		c.log.Info("Running auto-submit job")
		err := c.student.AutoSubmitAssessment()
		if err != nil {
			c.log.Error("Failed to run auto-submit job", zap.Error(err))
		} else {
			c.log.Info("Auto-submit job completed successfully")
		}
	}))
	if err != nil {
		return
	}

	c.cron.Start()
	c.log.Info(fmt.Sprintf("Auto-submit job started with ID: %d", job))
}
