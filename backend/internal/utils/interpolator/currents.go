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
		for _, c := range uCurrentsData {
			i.logger.Info("Before", "u_currents", c.UCurrent)
		}
		i.interpolateLinearyDataRow(interpolableUCurrentsDataSlice)
		for _, c := range uCurrentsData {
			i.logger.Info("After", "u_currents", c.UCurrent)
		}
		i.db.UpdateUCurrentsData(ctx, uCurrentsData)

		vCurrentsData, err := i.db.GetVCurrentsDataAtLocation(ctx, p)
		if err != nil {
			i.logger.Error("error geting v_currents data at location", "loc", p, "err", err)
		}
		interpolableVCurrentsDataSlice := make([]InterpolatableData, len(vCurrentsData))
		for i := range vCurrentsData {
			interpolableVCurrentsDataSlice[i] = &vCurrentsData[i]
		}
		for _, c := range vCurrentsData {
			i.logger.Info("Before", "v_currents", c.VCurrent)
		}
		i.interpolateLinearyDataRow(interpolableVCurrentsDataSlice)
		for _, c := range vCurrentsData {
			i.logger.Info("After", "v_currents", c.VCurrent)
		}
		i.db.UpdateVCurrentsData(ctx, vCurrentsData)
	}
	i.logger.Info("Interpolation of data based on time completed")
	return nil
}

func (i *Interpolator) RunCurrentsInterpolationBasedOnArea(ctx context.Context) error {
	//TODO:
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
