package distribution_test

import (
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/analytics/distribution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistogram(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	binEdges, binCounts, err := distribution.Histogram(data, 4)
	require.NoError(t, err)
	require.Len(t, binEdges, 5)
	require.Len(t, binCounts, 4)

	total := 0
	for _, c := range binCounts {
		total += c
	}
	assert.Equal(t, len(data), total)
}

func TestFormatHistogramForChartJS(t *testing.T) {
	binEdges := []float64{0, 1, 2}
	binCounts := []int{3, 7}
	labels, counts := distribution.FormatHistogramForChartJS(binEdges, binCounts)
	expectedLabels := []string{"[0.00, 1.00]", "[1.00, 2.00]"}
	require.Len(t, labels, len(expectedLabels))
	require.Len(t, counts, 2)

	for i, label := range labels {
		assert.Equal(t, expectedLabels[i], label)
	}
}

func TestKDEApproximate(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	xs, ys, err := distribution.KDEApproximate(data, 10, 0.5)
	require.NoError(t, err)
	require.Len(t, xs, 10)
	require.Len(t, ys, 10)

	for _, y := range ys {
		assert.GreaterOrEqual(t, y, 0.0)
	}
}

func TestFormatKDEForChartJS(t *testing.T) {
	xs := []float64{1.0, 2.0}
	ys := []float64{0.1, 0.2}
	labels, densities := distribution.FormatKDEForChartJS(xs, ys)
	expectedLabels := []string{"1.00", "2.00"}
	require.Len(t, labels, len(expectedLabels))

	for i, label := range labels {
		assert.Equal(t, expectedLabels[i], label)
	}
	assert.InDeltaSlice(t, ys, densities, 1e-9)
}
