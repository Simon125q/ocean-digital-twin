package interpolator

import (
	"context"
	"log/slog"
	"ocean-digital-twin/internal/database"
)

type Interpolator struct {
	db     database.Service
	logger *slog.Logger
}

func NewInterpolator(db database.Service, logger *slog.Logger) *Interpolator {
	return &Interpolator{
		db:     db,
		logger: logger,
	}
}

func (i *Interpolator) Run(ctx context.Context) error {
	i.logger.Info("Starting interpolation of data")

	//TODO: implement logic
	// endTime := time.Now().UTC()
	// startTime := endTime.Add(-7 * 24 * time.Hour)
	// minLat := 40.83
	// minLon := 1.10
	// maxLat := 41.26
	// maxLon := 2.53
	points, err := i.db.GetAllChlorophyllLocations(ctx)
	if err != nil {
		i.logger.Error("error geting chlor locations", "err", err)
	}
	// for _, c := range points {
	// 	i.logger.Info("chlor_a", "lat", c.Lat(), "lon", c.Lon())
	//
	// }
	chlorData, err := i.db.GetChlorophyllDataAtLocation(ctx, points[1])
	if err != nil {
		i.logger.Error("error geting chlor data at location", "err", err)
	}
	for _, d := range chlorData {
		i.logger.Info("chlor data", "time", d.MeasurementTime, "val", d.ChlorophyllA)
	}
	i.logger.Info("Success getting location points", "count", len(points))
	i.logger.Info("Interpolation of data completed")
	return nil
}
