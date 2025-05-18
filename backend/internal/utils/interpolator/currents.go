package interpolator

import "context"

func (i *Interpolator) RunLinearCurrentsInterpolationBasedOnTime(ctx context.Context) error {
	i.logger.Info("Starting interpolation of data based on time")

	points, err := i.db.GetAllCurrentsLocations(ctx)
	if err != nil {
		i.logger.Error("error geting currents locations", "err", err)
		return err
	}
	i.logger.Info("Success getting location points", "count", len(points))
	for _, p := range points {
		uCurrentsData, err := i.db.GetUCurrentsDataAtLocation(ctx, p)
		if err != nil {
			i.logger.Error("error geting u_currents data at location", "loc", p, "err", err)
		}
		interpolableUCurrentsDataSlice := make([]InterpolatableData, len(uCurrentsData))
		for i := range uCurrentsData {
			interpolableUCurrentsDataSlice[i] = &uCurrentsData[i]
		}
		i.interpolateLinearyDataRow(interpolableUCurrentsDataSlice)
		i.db.UpdateUCurrentsData(ctx, uCurrentsData)

		vCurrentsData, err := i.db.GetVCurrentsDataAtLocation(ctx, p)
		if err != nil {
			i.logger.Error("error geting v_currents data at location", "loc", p, "err", err)
		}
		interpolableVCurrentsDataSlice := make([]InterpolatableData, len(vCurrentsData))
		for i := range vCurrentsData {
			interpolableVCurrentsDataSlice[i] = &vCurrentsData[i]
		}
		i.interpolateLinearyDataRow(interpolableVCurrentsDataSlice)
		i.db.UpdateVCurrentsData(ctx, vCurrentsData)
	}
	i.logger.Info("Interpolation of data based on time completed")
	return nil
}

func (i *Interpolator) RunCurrentsInterpolationBasedOnArea(ctx context.Context) error {
	//TODO:
	i.logger.Info("Starting interpolation of data area")

	timestamps, err := i.db.GetAllCurrentsTimestamps(ctx)
	if err != nil {
		i.logger.Error("error geting currents timestamps", "err", err)
		return err
	}
	i.logger.Info("Success getting timestamps", "count", len(timestamps))
	for _, t := range timestamps {
		vCurrentsData, err := i.db.GetVCurrentDataAtTimestamp(ctx, t)
		if err != nil {
			i.logger.Error("error geting v_current data at timestamp", "time", t, "err", err)
		}

		interpolableVCurrentDataSlice := make([][]InterpolatableData, len(vCurrentsData))
		for i := range vCurrentsData {
			interpolableVCurrentDataSlice[i] = make([]InterpolatableData, len(vCurrentsData[i]))
		}

		for row := range vCurrentsData {
			for col := range vCurrentsData[row] {
				interpolableVCurrentDataSlice[row][col] = &vCurrentsData[row][col]
			}
		}
		i.interpolateDataArea(interpolableVCurrentDataSlice)
		for row := range vCurrentsData {
			i.db.UpdateVCurrentsData(ctx, vCurrentsData[row])
		}

		uCurrentsData, err := i.db.GetUCurrentDataAtTimestamp(ctx, t)
		if err != nil {
			i.logger.Error("error geting u_current data at timestamp", "time", t, "err", err)
		}

		interpolableUCurrentDataSlice := make([][]InterpolatableData, len(uCurrentsData))
		for i := range uCurrentsData {
			interpolableUCurrentDataSlice[i] = make([]InterpolatableData, len(uCurrentsData[i]))
		}

		for row := range uCurrentsData {
			for col := range uCurrentsData[row] {
				interpolableUCurrentDataSlice[row][col] = &uCurrentsData[row][col]
			}
		}
		i.interpolateDataArea(interpolableUCurrentDataSlice)
		for row := range uCurrentsData {
			i.db.UpdateUCurrentsData(ctx, uCurrentsData[row])
		}
	}
	i.logger.Info("Interpolation of data based on area completed")
	return nil
}
