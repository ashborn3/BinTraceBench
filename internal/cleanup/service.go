package cleanup

import (
	"context"
	"time"

	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/ashborn3/BinTraceBench/pkg/logging"
)

type Service struct {
	db       database.Database
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewService(db database.Database, interval time.Duration) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		db:       db,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *Service) Start() {
	go s.run()
}

func (s *Service) Stop() {
	s.cancel()
}

func (s *Service) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupExpiredSessions()
		}
	}
}

func (s *Service) cleanupExpiredSessions() {
	err := s.db.DeleteExpiredSessions()
	if err != nil {
		logging.Error("Failed to cleanup expired sessions", "error", err)
	} else {
		logging.Debug("Cleaned up expired sessions")
	}
}
