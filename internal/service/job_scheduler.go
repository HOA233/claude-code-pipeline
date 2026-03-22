package service

import (
	"context"
	"time"

	"github.com/company/claude-pipeline/internal/repository"
	"github.com/company/claude-pipeline/pkg/logger"
)

// Scheduler handles scheduled job execution
type Scheduler struct {
	redis  *repository.RedisClient
	jobSvc *ScheduledJobService
	stopCh chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(redis *repository.RedisClient, jobSvc *ScheduledJobService) *Scheduler {
	return &Scheduler{
		redis:  redis,
		jobSvc: jobSvc,
		stopCh: make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) {
	logger.Info("Scheduler started")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			logger.Info("Scheduler stopped")
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndRunJobs(ctx)
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// checkAndRunJobs checks all scheduled jobs and triggers due ones
func (s *Scheduler) checkAndRunJobs(ctx context.Context) {
	jobs, err := s.jobSvc.ListScheduledJobs(ctx, "")
	if err != nil {
		logger.Error("Failed to list scheduled jobs: ", err)
		return
	}

	now := time.Now()
	for _, job := range jobs {
		if !job.Enabled {
			continue
		}

		if job.NextRun != nil && (job.NextRun.Before(now) || job.NextRun.Equal(now)) {
			go s.triggerJob(ctx, job.ID)
		}
	}
}

// triggerJob triggers a scheduled job
func (s *Scheduler) triggerJob(ctx context.Context, jobID string) {
	_, err := s.jobSvc.TriggerJob(ctx, jobID)
	if err != nil {
		logger.Error("Failed to trigger job: ", err)
	}
}