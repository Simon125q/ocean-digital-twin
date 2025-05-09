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

	i.logger.Info("Interpolation of data completed")
	return nil
}
