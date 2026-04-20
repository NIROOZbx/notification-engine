package scheduler

import (
	"context"
	"time"

	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/NIROOZbx/notification-engine/engine/notification/queue"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/rs/zerolog"
)

type Scheduler struct {
	repo     repositories.SchedulerRepo
	producer core.Producer
	log      zerolog.Logger
	interval time.Duration
}

func NewScheduler(producer core.Producer, log zerolog.Logger, repo repositories.SchedulerRepo, interval time.Duration) *Scheduler {
	return &Scheduler{producer: producer, log: log, repo: repo, interval: interval}
}

func (s *Scheduler) pollScheduled(ctx context.Context) {
	s.log.Debug().Msg("SCHEDULER_TICK: Checking for due notifications")
	logs, err := s.repo.GetDueScheduledNotifications(ctx, 10)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to fetch scheduled notifications")
		return
	}
	s.poll(ctx, logs)
}

func (s *Scheduler) pollRetry(ctx context.Context) {
	s.log.Debug().Msg("RETRY_TICK: Checking for due retries")
	logs, err := s.repo.GetDueRetryNotifications(ctx, 10)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to fetch retry notifications")
		return
	}
	s.poll(ctx, logs)
}

func (s *Scheduler) poll(ctx context.Context, logs []*core.NotificationLog) {
	for _, log := range logs {
		event := models.NotificationEvent{
			NotificationLogID: log.ID,
			WorkspaceID:       log.WorkspaceID,
			EnvironmentID:     log.EnvironmentID,
			Channel:           log.Channel,
			Recipient:         log.Recipient,
			AttemptNumber:     log.AttemptCount,
			Data:              log.TriggerData,
		}

		if err := s.repo.UpdateNotificationStatus(ctx, log.ID, "processing"); err != nil {
			s.log.Error().Err(err).Str("log_id", log.ID).Msg("failed to claim notification")
			continue
		}

		topic, err := queue.TopicByChannel(log.Channel)
		if err != nil {
			s.log.Error().Err(err).Str("log_id", log.ID).Msg("unsupported channel, marking as failed")
			_ = s.repo.UpdateNotificationStatus(ctx, log.ID, "failed")
			continue
		}

		if err := s.producer.Publish(ctx, topic, event); err != nil {
			s.log.Error().Err(err).Str("log_id", log.ID).Msg("failed to publish to kafka, rolling back status")

			if err := s.repo.UpdateNotificationStatus(ctx, log.ID, log.Status); err != nil {
				s.log.Error().Err(err).Str("log_id", log.ID).Msg("critical: failed to rollback log status")
			}
			continue
		}

		s.log.Debug().Str("log_id", log.ID).Msg("successfully injected scheduled notification into pipeline")
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	s.log.Info().Dur("interval", s.interval).Msg("scheduler started")

	for {
		select {
		case <-ticker.C:
			s.Run(ctx)
		case <-ctx.Done():
			s.log.Info().Msg("scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.pollScheduled(ctx)
	s.pollRetry(ctx)
}
