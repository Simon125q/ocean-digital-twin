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

	points, err := i.db.GetAllChlorophyllLocations(ctx)
	if err != nil {
		i.logger.Error("error geting chlor locations", "err", err)
		return err
	}
	i.logger.Info("Success getting location points", "count", len(points))
	for _, p := range points {
		chlorData, err := i.db.GetChlorophyllDataAtLocation(ctx, p)
		if err != nil {
			i.logger.Error("error geting chlor data at location", "loc", p, "err", err)
		}
		interpolableDataSlice := make([]InterpolableData, len(chlorData))
		for i := range chlorData {
			interpolableDataSlice[i] = &chlorData[i]
		}
		i.interpolateLinearyDataRow(interpolableDataSlice)
		i.db.UpdateChlorophyllData(ctx, chlorData)
	}
	i.logger.Info("Interpolation of data completed")
	return nil
}

func (ip *Interpolator) interpolateLinearyDataRow(data []InterpolableData) {
	if len(data) < 3 {
		return
	}
	for i := 0; i < len(data); i++ {
		if math.IsNaN(float64(data[i].Value())) {
			// If only 1 value in a row is missing fill it with the average of surrounding values
			if i > 0 && i < len(data)-1 &&
				!math.IsNaN(float64(data[i-1].Value())) && !math.IsNaN(float64(data[i+1].Value())) {
				data[i].SetValue((data[i-1].Value() + data[i+1].Value()) / 2.0)
				continue
			}

			// If more than 1 value in a row is missing interpolate the missing values
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
						data[l].SetValue(startValue + (endValue-startValue)*float32(step)/float32(gapLength))
					}
					i = gapEndIndex - 1
					continue
				}
			}
		}
	}
	return
}
