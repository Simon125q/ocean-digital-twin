package scheduler

import (
	"context"
	"log/slog"
	"ocean-digital-twin/internal/database"
	"ocean-digital-twin/internal/utils/erddap"
	"ocean-digital-twin/internal/utils/interpolator"
	"time"
)

type Updater struct {
	db           database.Service
	downloader   *erddap.Downloader
	interpolator *interpolator.Interpolator
	logger       *slog.Logger
	interval     time.Duration
	minLat       float64
	maxLat       float64
	minLon       float64
	maxLon       float64
}

func NewUpdater(
	db database.Service,
	logger *slog.Logger,
	interval time.Duration,
	minLat, minLon, maxLat, maxLon float64,
) *Updater {
	return &Updater{
		db:           db,
		downloader:   erddap.NewDownloader(logger, minLat, minLon, maxLat, maxLon),
		interpolator: interpolator.NewInterpolator(db, logger),
		logger:       logger,
		interval:     interval,
		minLat:       minLat,
		minLon:       minLon,
		maxLat:       maxLat,
		maxLon:       maxLon,
	}
}

func (u *Updater) Start(ctx context.Context) {
	u.update(ctx)

	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			u.update(ctx)
		case <-ctx.Done():
			u.logger.Info("Updater Stopped")
			return
		}
	}
}

func (u *Updater) update(ctx context.Context) {
	u.updateChlorophyllData(ctx)

	u.interpolator.Run(ctx)
}
