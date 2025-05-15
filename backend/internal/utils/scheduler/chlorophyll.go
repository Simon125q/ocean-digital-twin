package scheduler

import (
	"context"
	"ocean-digital-twin/internal/utils/erddap"
	"time"
)

func (u *Updater) updateChlorophyllData(ctx context.Context) {
	u.logger.Info("Starting chlorophyll data update")

	latestTime, err := u.db.GetLatestChlorophyllTimestamp(ctx)
	if err != nil {
		u.logger.Error("Failed to get latest timestamp", "error", err)
	}
	startTime := latestTime
	// if start time is older than 30 days set it to 30 days
	if time.Since(startTime) > 30*24*time.Hour {
		startTime = time.Now().UTC().Add(-30 * 24 * time.Hour)
	}

	endTime, err := u.downloader.GetLatestDataTime(ctx, erddap.ChlorDatasetID)
	if err != nil {
		u.logger.Error("Coudnt get latest time for chlorophyll from ERDDAP")
		return
	}

	if !startTime.Before(endTime) {
		u.logger.Info("Chlorophyll - Latest timestamp of data in db is after or equal the latest timestamp available in erddap - no data to update")
		return
	}

	data, err := u.downloader.DownloadChlorophyllData(ctx, startTime, endTime)
	if err != nil {
		u.logger.Error("Failed to download chlorophyll data", "err", err)
		return
	}

	if len(data) == 0 {
		u.logger.Info("No new chlorophyll data available")
		return
	}

	if err := u.db.SaveChlorophyllData(ctx, data); err != nil {
		u.logger.Error("Failed to save chlorophyll data", "err", err)
		return
	}
	if err := u.db.SaveChlorophyllDataRaw(ctx, data); err != nil {
		u.logger.Error("Failed to save chlorophyll data", "err", err)
		return
	}
	u.logger.Info("Chlorophyll data update completed", "updated_points", len(data))
}
