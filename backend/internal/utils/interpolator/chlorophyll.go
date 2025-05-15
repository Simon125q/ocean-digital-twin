package interpolator

import "context"

func (i *Interpolator) RunLinearChlorophyllInterpolationBasedOnTime(ctx context.Context) error {
	i.logger.Info("Starting interpolation of data based on time")

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
		interpolableDataSlice := make([]InterpolatableData, len(chlorData))
		for i := range chlorData {
			interpolableDataSlice[i] = &chlorData[i]
		}
		i.interpolateLinearyDataRow(interpolableDataSlice)
		i.db.UpdateChlorophyllData(ctx, chlorData)
	}
	i.logger.Info("Interpolation of data based on time completed")
	return nil
}

func (i *Interpolator) RunChlorophyllInterpolationBasedOnArea(ctx context.Context) error {
	i.logger.Info("Starting interpolation of data area")

	timestamps, err := i.db.GetAllChlorophyllTimestamps(ctx)
	if err != nil {
		i.logger.Error("error geting chlor timestamps", "err", err)
		return err
	}
	i.logger.Info("Success getting timestamps", "count", len(timestamps))
	for _, t := range timestamps {
		chlorData, err := i.db.GetChlorophyllDataAtTimestamp(ctx, t)
		if err != nil {
			i.logger.Error("error geting chlor data at timestamp", "time", t, "err", err)
		}

		interpolableDataSlice := make([][]InterpolatableData, len(chlorData))
		for i := range chlorData {
			interpolableDataSlice[i] = make([]InterpolatableData, len(chlorData[i]))
		}

		for row := range chlorData {
			for col := range chlorData[row] {
				interpolableDataSlice[row][col] = &chlorData[row][col]
			}
		}
		i.interpolateDataArea(interpolableDataSlice)
		for row := range chlorData {
			i.db.UpdateChlorophyllData(ctx, chlorData[row])
		}
	}
	i.logger.Info("Interpolation of data based on area completed")
	return nil
}
