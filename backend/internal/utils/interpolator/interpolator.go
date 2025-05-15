package interpolator

import (
	"context"
	"log/slog"
	"math"
	"ocean-digital-twin/internal/database"
)

type InterpolatableData interface {
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

func (i *Interpolator) RunChlorophyllInterpolation(ctx context.Context) error {
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
		interpolableDataSlice := make([]InterpolatableData, len(chlorData))
		for i := range chlorData {
			interpolableDataSlice[i] = &chlorData[i]
		}
		i.interpolateLinearyDataRow(interpolableDataSlice)
		i.db.UpdateChlorophyllData(ctx, chlorData)
	}
	i.logger.Info("Interpolation of data completed")
	return nil
}

func (ip *Interpolator) interpolateLinearyDataRow(data []InterpolatableData) {
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

func (ip *Interpolator) interpolateDataArea(data [][]InterpolatableData) [][]InterpolatableData {
	if len(data) == 0 || len(data[0]) == 0 {
		return data
	}

	rows := len(data)
	cols := len(data[0])

	filledData := make([][]InterpolatableData, rows)
	for i := range filledData {
		filledData[i] = make([]InterpolatableData, cols)
		copy(filledData[i], data[i])
	}

	visited := make([][]bool, rows)
	for i := range visited {
		visited[i] = make([]bool, cols)
	}

	// Directions for exploring neighbors (including diagonals)
	dr := []int{-1, -1, -1, 0, 0, 1, 1, 1}
	dc := []int{-1, 0, 1, -1, 1, -1, 0, 1}

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if math.IsNaN(float64(filledData[r][c].Value())) && !visited[r][c] {
				// Found an unvisited NaN, start exploring the group
				groupCoords := make([][2]int, 0)
				neighborValues := make([]float32, 0)
				isSurrounded := true

				// Use a queue for BFS
				queue := [][2]int{{r, c}}
				visited[r][c] = true
				groupCoords = append(groupCoords, [2]int{r, c})

				for len(queue) > 0 {
					currR, currC := queue[0][0], queue[0][1]
					queue = queue[1:]

					// Explore neighbors
					for i := 0; i < 8; i++ {
						nR, nC := currR+dr[i], currC+dc[i]

						// Check bounds
						if nR < 0 || nR >= rows || nC < 0 || nC >= cols {
							isSurrounded = false
							continue
						}

						if math.IsNaN(float64(filledData[nR][nC].Value())) {
							if !visited[nR][nC] {
								visited[nR][nC] = true
								queue = append(queue, [2]int{nR, nC})
								groupCoords = append(groupCoords, [2]int{nR, nC})
							}
						} else {
							// Found a non-NaN neighbor, add its value
							neighborValues = append(neighborValues, filledData[nR][nC].Value())
						}
					}
				}

				// After exploring the group, check if it's surrounded and has neighbors
				if isSurrounded && len(neighborValues) > 0 {
					sum := float32(0.0)
					for _, val := range neighborValues {
						sum += val
					}
					average := sum / float32(len(neighborValues))

					for _, coord := range groupCoords {
						filledData[coord[0]][coord[1]].SetValue(average)
					}
				}
			}
		}
	}

	return filledData
}
