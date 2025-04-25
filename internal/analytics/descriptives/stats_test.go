package descriptives_test

import (
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/analytics/descriptives"
	"github.com/stretchr/testify/assert"
)

func TestMean(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"normal data", []float64{1, 2, 3, 4, 5}, 3, false},
		{"single value", []float64{42}, 42, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Mean(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InEpsilon(t, tc.expected, result, 1e-9)
			}
		})
	}

}

func TestMedian(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"odd", []float64{3, 1, 2}, 2, false},
		{"even", []float64{1, 2, 3, 4}, 2.5, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Median(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InEpsilon(t, tc.expected, result, 1e-9)
			}
		})
	}
}

func TestMode(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float64
		wantErr  bool
	}{
		{"single mode", []float64{1, 2, 2, 3}, []float64{2}, false},
		{"multi mode", []float64{1, 1, 2, 2}, []float64{1, 2}, false},
		{"empty", []float64{}, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Mode(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, result)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"standard", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 2.138, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.StdDev(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InEpsilon(t, tc.expected, result, 0.001)
			}
		})
	}
}

func TestVariance(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"standard", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 4.571, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Variance(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InEpsilon(t, tc.expected, result, 0.001)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"standard", []float64{5, 1, 8, 3}, 1, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Min(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"standard", []float64{5, 1, 8, 3}, 8, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Max(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"standard", []float64{1, 5, 9}, 8, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Range(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64
		wantErr  bool
	}{
		{"standard", []float64{1, 2, 3}, 6, false},
		{"empty", []float64{}, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := descriptives.Sum(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestCount(t *testing.T) {
	assert.Equal(t, 3, descriptives.Count([]float64{1, 2, 3}))
	assert.Equal(t, 0, descriptives.Count([]float64{}))
}
