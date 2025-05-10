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

// Helper function to check if two float32 slices are approximately equal (same as before)
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

// Make sure your interpolateDataRow function signature is updated like this:
// func interpolateDataRow(data []InterpolatableData) { ... }
