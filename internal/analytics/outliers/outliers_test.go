package outliers_test

import (
	"fmt"
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/analytics/outliers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZScoreOutliers(t *testing.T) {
	data := []float64{1, 2, 3, 4, 100}
	threshold := 1.5
	outliers, err := outliers.ZScoreOutliers(data, threshold)
	require.NoError(t, err)
	assert.ElementsMatch(t, outliers, []int{4}, "Outlier index should be 4")
}

func TestIQROutliersNone(t *testing.T) {
	data := []float64{1, 2, 3, 4, 100}
	outliers, err := outliers.IQROutliers(data)
	require.NoError(t, err)

	// Since no outliers should be detected based on the current thresholds
	assert.ElementsMatch(t, outliers, []int{}, "No outliers should be detected")
}

func TestIQROutliersWithOutlier(t *testing.T) {
	// Adding an outlier, say 1000, which is greater than the upper threshold of 247
	data := []float64{1, 2, 3, 4, 100, 1000}
	outliers, err := outliers.IQROutliers(data)
	require.NoError(t, err)

	// Debugging line to print outliers
	fmt.Println("Detected Outliers: ", outliers)

	// The index 5 (corresponding to the value 1000) should be flagged as an outlier
	assert.ElementsMatch(t, outliers, []int{5}, "Outlier index should be 5 (corresponding to 1000)")
}

func TestBoxPlotData(t *testing.T) {
	data := []float64{1, 2, 3, 4, 100}
	Q1, Q3, IQR, lowerOutlier, upperOutlier, err := outliers.BoxPlotData(data)
	require.NoError(t, err)

	assert.InDelta(t, 1.5, Q1, 0.01, "Q1 should be close to 1.5")
	assert.InDelta(t, 52.0, Q3, 0.01, "Q3 should be close to 52.0")
	assert.InDelta(t, 50.5, IQR, 0.01, "IQR should be close to 50.5")
	assert.InDelta(t, -74.25, lowerOutlier, 0.01, "Lower outlier threshold should be close to -74.25")
	assert.InDelta(t, 127.75, upperOutlier, 0.01, "Upper outlier threshold should be close to 127.75")
}

func TestFormatBoxPlotForChartJS(t *testing.T) {
	Q1, Q3, lowerOutlier, upperOutlier := 2.0, 4.0, -1.0, 7.0
	labels, values := outliers.FormatBoxPlotForChartJS(Q1, Q3, lowerOutlier, upperOutlier)
	expectedLabels := []string{"Q1", "Q3", "Lower Outlier", "Upper Outlier"}
	expectedValues := []float64{2.0, 4.0, -1.0, 7.0}

	assert.Equal(t, expectedLabels, labels, "Labels should match expected labels")
	assert.Equal(t, expectedValues, values, "Values should match expected values")
}

func TestQuantiles(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	Q1, Q2, Q3, err := outliers.Quantiles(data)
	require.NoError(t, err)
	assert.InDelta(t, 2.5, Q1, 0.01, "Q1 should be close to 3.0")
	assert.InDelta(t, 5.0, Q2, 0.01, "Q2 (Median) should be close to 5.0")
	assert.InDelta(t, 7.5, Q3, 0.01, "Q3 should be close to 7.0")
}
