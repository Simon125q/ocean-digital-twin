package interpolator

import (
	"context"
	"log/slog"
	"math"
	"ocean-digital-twin/internal/database"
)

type InterpolableData interface {
	Value() float32
	SetValue(float32)
}

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
		i.logger.Info("chlor data before", "time", d.MeasurementTime, "id", d.ID, "val", d.ChlorophyllA)
	}
	middleDataSlice := make([]InterpolableData, len(chlorData))
	for i := range chlorData {
		middleDataSlice[i] = &chlorData[i]
	}
	i.interpolateLinearyDataRow(middleDataSlice)
	for _, d := range chlorData {
		i.logger.Info("chlor data after", "time", d.MeasurementTime, "id", d.ID, "val", d.ChlorophyllA)
	}
	i.logger.Info("Success getting location points", "count", len(points))
	i.logger.Info("Interpolation of data completed")
	return nil
}

func (ip *Interpolator) interpolateLinearyDataRow(data []InterpolableData) {
	if len(data) < 3 {
		return
	}
	for i := 0; i < len(data); i++ {
		if math.IsNaN(float64(data[i].Value())) {
			if i > 0 && i < len(data)-1 &&
				!math.IsNaN(float64(data[i-1].Value())) && !math.IsNaN(float64(data[i+1].Value())) {
				data[i].SetValue((data[i-1].Value() + data[i+1].Value()) / 2.0)
				continue
			}

			if i > 0 && !math.IsNaN(float64(data[i-1].Value())) {
				gapEndIndex := -1
				for k := i + 1; k < len(data); k++ {
					if !math.IsNaN(float64(data[k].Value())) {
						gapEndIndex = k
						break
					}
				}
				if gapEndIndex != -1 && gapEndIndex > i {
					startValue := data[i-1].Value()
					endValue := data[gapEndIndex].Value()
					gapLength := gapEndIndex - (i - 1)
					for l := i; l < gapEndIndex; l++ {
						step := l - (i - 1)
						ip.logger.Info("Setting new vals", "val before", data[l].Value())
						data[l].SetValue(startValue + (endValue-startValue)*float32(step)/float32(gapLength))
						ip.logger.Info("Setting new vals", "val after", data[l].Value())
					}
					i = gapEndIndex - 1
					continue
				}
			}
		}
	}
	return
}
