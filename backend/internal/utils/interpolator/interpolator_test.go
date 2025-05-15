package interpolator

import (
	"log/slog"
	"math"
	"testing"
)

// Mock implementation of InterpolatableData for testing
type mockInterpolatableData struct {
	val float32
}

func (m *mockInterpolatableData) Value() float32 {
	return m.val
}

func (m *mockInterpolatableData) SetValue(v float32) {
	m.val = v
}

// Helper function to create a slice of mockInterpolatableData
func newMockDataSlice(values []float32) []InterpolatableData {
	slice := make([]InterpolatableData, len(values))
	for i, v := range values {
		slice[i] = &mockInterpolatableData{val: v}
	}
	return slice
}

// Helper function to extract float32 values from a slice of InterpolatableData
func extractValues(dataSlice []InterpolatableData) []float32 {
	values := make([]float32, len(dataSlice))
	for i, item := range dataSlice {
		values[i] = item.Value()
	}
	return values
}

func newMock2DDataSlice(values [][]float32) [][]InterpolatableData {
	slice := make([][]InterpolatableData, len(values))
	for i, row := range values {
		slice[i] = make([]InterpolatableData, len(row))
		for j, v := range row {
			slice[i][j] = &mockInterpolatableData{val: v}
		}
	}
	return slice
}

func extractValuesFrom2DSlice(dataSlice [][]InterpolatableData) [][]float32 {
	values := make([][]float32, len(dataSlice))
	for i, row := range dataSlice {
		values[i] = make([]float32, len(row))
		for j, item := range row {
			values[i][j] = item.Value()
		}
	}
	return values
}

func areFloat32SlicesEqual(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.IsNaN(float64(a[i])) != math.IsNaN(float64(b[i])) {
			return false
		}
		if !math.IsNaN(float64(a[i])) {
			const epsilon = 1e-6
			if math.Abs(float64(a[i])-float64(b[i])) > epsilon {
				return false
			}
		}
	}
	return true
}

func are2DFloat32SlicesEqual(s1, s2 [][]float32) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if len(s1[i]) != len(s2[i]) {
			return false
		}
		for j := range s1[i] {
			// Special handling for NaN: math.IsNaN(x) == math.IsNaN(y) should be true if both are NaN
			if math.IsNaN(float64(s1[i][j])) && math.IsNaN(float64(s2[i][j])) {
				continue
			}
			if s1[i][j] != s2[i][j] {
				return false
			}
		}
	}
	return true
}

func Test_interpolateDataRow_WithInterface(t *testing.T) {
	// Define NaN for convenience in test cases
	nan := float32(math.NaN())

	tests := []struct {
		name     string
		input    []float32 // Use float32 for input definition simplicity
		expected []float32
	}{
		{
			name:     "Empty slice",
			input:    []float32{},
			expected: []float32{},
		},
		{
			name:     "Slice with one element",
			input:    []float32{1.0},
			expected: []float32{1.0},
		},
		{
			name:     "Slice with two elements",
			input:    []float32{1.0, 2.0},
			expected: []float32{1.0, 2.0},
		},
		{
			name:     "Slice with two elements and NaN",
			input:    []float32{1.0, nan},
			expected: []float32{1.0, nan},
		},
		{
			name:     "Slice with three elements, middle NaN, direct neighbors valid",
			input:    []float32{1.0, nan, 2.0},
			expected: []float32{1.0, 1.5, 2.0}, // (1.0 + 2.0) / 2 = 1.5
		},
		{
			name:     "Slice with three elements, first NaN",
			input:    []float32{nan, 1.0, 2.0},
			expected: []float32{nan, 1.0, 2.0}, // Gap at the beginning, unbounded
		},
		{
			name:     "Slice with three elements, last NaN",
			input:    []float32{1.0, 2.0, nan},
			expected: []float32{1.0, 2.0, nan}, // Gap at the end, unbounded
		},
		{
			name:     "Slice with multiple NaNs, direct neighbors valid",
			input:    []float32{1.0, nan, 2.0, nan, 3.0},
			expected: []float32{1.0, 1.5, 2.0, 2.5, 3.0}, // (1.0+2.0)/2=1.5, (2.0+3.0)/2=2.5
		},
		{
			name:     "Linear interpolation, one NaN",
			input:    []float32{0.0, nan, 2.0},
			expected: []float32{0.0, 1.0, 2.0}, // 0.0 + (2.0-0.0) * 1/2 = 1.0
		},
		{
			name:     "Linear interpolation, two NaNs",
			input:    []float32{0.0, nan, nan, 3.0},
			expected: []float32{0.0, 1.0, 2.0, 3.0}, // 0.0 + (3.0-0.0) * 1/3 = 1.0, 0.0 + (3.0-0.0) * 2/3 = 2.0
		},
		{
			name:     "Linear interpolation, multiple NaNs",
			input:    []float32{1.0, nan, nan, nan, 5.0},
			expected: []float32{1.0, 2.0, 3.0, 4.0, 5.0}, // 1 + (5-1)*1/4=2, 1+(5-1)*2/4=3, 1+(5-1)*3/4=4
		},
		{
			name:     "Linear interpolation with negative values",
			input:    []float32{-1.0, nan, nan, 2.0},
			expected: []float32{-1.0, 0.0, 1.0, 2.0}, // -1 + (2-(-1))*1/3=0, -1+(2-(-1))*2/3=1
		},
		{
			name:     "Combined simple and linear interpolation",
			input:    []float32{1.0, nan, 3.0, nan, nan, 9.0, nan, 11.0},
			expected: []float32{1.0, 2.0, 3.0, 5.0, 7.0, 9.0, 10.0, 11.0},
			// (1+3)/2=2
			// 3 + (9-3)*1/3 = 3 + 6*1/3 = 3+2=5
			// 3 + (9-3)*2/3 = 3 + 6*2/3 = 3+4=7
			// (9+11)/2=10
		},
		{
			name:     "Gap at the beginning, no interpolation",
			input:    []float32{nan, nan, 3.0, 4.0},
			expected: []float32{nan, nan, 3.0, 4.0},
		},
		{
			name:     "Gap at the end, no interpolation",
			input:    []float32{1.0, 2.0, nan, nan},
			expected: []float32{1.0, 2.0, nan, nan},
		},
		{
			name:     "Gap in the middle, not bounded at the end",
			input:    []float32{1.0, nan, nan},
			expected: []float32{1.0, nan, nan},
		},
		{
			name:     "Gap in the middle, not bounded at the beginning",
			input:    []float32{nan, nan, 3.0},
			expected: []float32{nan, nan, 3.0},
		},
		{
			name:     "All NaNs",
			input:    []float32{nan, nan, nan},
			expected: []float32{nan, nan, nan},
		},
		{
			name:     "Valid data only",
			input:    []float32{1.0, 2.0, 3.0, 4.0},
			expected: []float32{1.0, 2.0, 3.0, 4.0},
		},
		{
			name:     "Single NaN in a longer series",
			input:    []float32{1.0, 2.0, nan, 4.0, 5.0},
			expected: []float32{1.0, 2.0, 3.0, 4.0, 5.0}, // Simple interpolation: (2.0 + 4.0) / 2 = 3.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputDataSlice := newMockDataSlice(tt.input)

			interpolator := NewInterpolator(nil, &slog.Logger{})
			interpolator.interpolateLinearyDataRow(inputDataSlice)

			resultValues := extractValues(inputDataSlice)

			if !areFloat32SlicesEqual(resultValues, tt.expected) {
				t.Errorf("interpolateDataRow() for input %v resulted in values %v, want %v",
					tt.input, resultValues, tt.expected)
			}
		})
	}
}

func Test_interpolateDataArea_WithInterface(t *testing.T) {
	nan := float32(math.NaN())

	tests := []struct {
		name     string
		input    [][]float32
		expected [][]float32
	}{
		{
			name: "Single NaN surrounded by data",
			input: [][]float32{
				{1, 1, 1},
				{1, nan, 1},
				{1, 1, 1},
			},
			expected: [][]float32{
				{1, 1, 1},
				{1, 1, 1}, // Average of 8 surrounding 1s is 1
				{1, 1, 1},
			},
		},
		{
			name: "Group of NaNs surrounded by data",
			input: [][]float32{
				{2, 2, 2, 2},
				{2, nan, nan, 2},
				{2, nan, nan, 2},
				{2, 2, 2, 2},
			},
			expected: [][]float32{
				{2, 2, 2, 2},
				{2, 2, 2, 2}, // Average of all surrounding 2s is 2
				{2, 2, 2, 2},
				{2, 2, 2, 2},
			},
		},
		{
			name: "NaN at the edge (should not be filled)",
			input: [][]float32{
				{1, 1, 1},
				{nan, 1, 1},
				{1, 1, 1},
			},
			expected: [][]float32{
				{1, 1, 1},
				{nan, 1, 1}, // Edge NaN remains
				{1, 1, 1},
			},
		},
		{
			name: "NaN group touching the edge (should not be filled)",
			input: [][]float32{
				{nan, nan, 1},
				{nan, 1, 1},
				{1, 1, 1},
			},
			expected: [][]float32{
				{nan, nan, 1}, // Group touching edge remains
				{nan, 1, 1},
				{1, 1, 1},
			},
		},
		{
			name: "Multiple separated surrounded NaN groups",
			input: [][]float32{
				{1, 1, 1, 5, 5, 5},
				{1, nan, 1, 5, nan, 5},
				{1, 1, 1, 5, 5, 5},
				{10, 10, 10, 20, 20, 20},
				{10, nan, 10, 20, nan, 20},
				{10, 10, 10, 20, 20, 20},
			},
			expected: [][]float32{
				{1, 1, 1, 5, 5, 5},
				{1, 1, 1, 5, 5, 5}, // First group filled with 1s avg
				{1, 1, 1, 5, 5, 5},
				{10, 10, 10, 20, 20, 20},
				{10, 10, 10, 20, 20, 20}, // Second group filled with 10s avg
				{10, 10, 10, 20, 20, 20}, // Third group filled with 20s avg
			},
		},
		{
			name: "NaN group with diagonal neighbors only (should be filled)",
			input: [][]float32{
				{1, 0, 1},
				{0, nan, 0},
				{1, 0, 1},
			},
			expected: [][]float32{
				{1, 0, 1},
				{0, 0.5, 0}, // Average of 1, 1, 0, 0, 1, 1, 0, 0 is 0.5
				{1, 0, 1},
			},
		},
		{
			name: "Array with no NaNs",
			input: [][]float32{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			expected: [][]float32{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
		},
		{
			name: "Array with all NaNs (should not be filled)",
			input: [][]float32{
				{nan, nan},
				{nan, nan},
			},
			expected: [][]float32{
				{nan, nan},
				{nan, nan},
			},
		},
		{
			name:     "Empty input slice",
			input:    [][]float32{},
			expected: [][]float32{},
		},
		{
			name:     "Slice with empty inner slices",
			input:    [][]float32{{}, {}},
			expected: [][]float32{{}, {}},
		},
		{
			name: "NaN surrounded by different values",
			input: [][]float32{
				{1, 2, 3},
				{4, nan, 5},
				{6, 7, 8},
			},
			expected: [][]float32{
				{1, 2, 3},
				{4, 4.5, 5}, // Avg of 1,2,3,4,5,6,7,8 = 36/8 = 4.5
				{6, 7, 8},
			},
		},
		{
			name: "Larger test case from description",
			input: [][]float32{
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 1, 1, 1, 1, 2, 2, 2},
				{2, 2, 2, 1, nan, nan, 1, 1, 2, 2},
				{2, 2, 2, 1, nan, nan, nan, 1, 2, 2},
				{2, 2, 2, 1, nan, nan, 1, 1, 2, 2},
				{2, 2, 2, 1, 1, 1, 1, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
			expected: [][]float32{
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 1, 1, 1, 1, 2, 2, 2},
				{2, 2, 2, 1, 1.0, 1.0, 1, 1, 2, 2}, // Average of surrounding 1s is 1.0
				{2, 2, 2, 1, 1.0, 1.0, 1.0, 1, 2, 2},
				{2, 2, 2, 1, 1.0, 1.0, 1, 1, 2, 2},
				{2, 2, 2, 1, 1, 1, 1, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
		},
		{
			name: "Larger test case from description",
			input: [][]float32{
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
				{nan, 1, 1, 1, 2, 2, 2, 2, 2, 2},
				{nan, nan, nan, 1, 2, 2, 2, 2, 2, 2},
				{nan, nan, nan, 1, 2, 2, 2, 2, 2, 2},
				{nan, nan, 1, 1, 2, 2, 2, 2, 2, 2},
				{1, 1, 1, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
			expected: [][]float32{
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
				{nan, 1, 1, 1, 2, 2, 2, 2, 2, 2}, // Group touches edge, remains NaN
				{nan, nan, nan, 1, 2, 2, 2, 2, 2, 2},
				{nan, nan, nan, 1, 2, 2, 2, 2, 2, 2},
				{nan, nan, 1, 1, 2, 2, 2, 2, 2, 2},
				{1, 1, 1, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				{2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
		},
		{
			name: "NaN group surrounded by different values",
			input: [][]float32{
				{1, 2, 3, 4},
				{5, nan, nan, 6},
				{7, nan, nan, 8},
				{9, 10, 11, 12},
			},
			expected: [][]float32{
				{1, 2, 3, 4},
				{5, 6.5, 6.5, 6},
				{7, 6.5, 6.5, 8},
				{9, 10, 11, 12},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputDataSlice := newMock2DDataSlice(tt.input)

			interpolator := NewInterpolator(nil, &slog.Logger{})
			result := interpolator.interpolateDataArea(inputDataSlice)

			resultValues := extractValuesFrom2DSlice(result)

			if !are2DFloat32SlicesEqual(resultValues, tt.expected) {
				t.Errorf("interpolateDataArea() for input %v resulted in values %v, want %v",
					tt.input, resultValues, tt.expected)
			}
		})
	}
}
