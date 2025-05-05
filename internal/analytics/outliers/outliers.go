package outliers

import (
	"fmt"
	"sort"

	"github.com/Bgoodwin24/insightforge/internal/analytics/descriptives"
)

func ZScoreOutliers(data []float64, threshold float64) ([]int, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	mean, err := descriptives.Mean(data)
	if err != nil {
		return nil, err
	}
	stddev, err := descriptives.StdDev(data)
	if err != nil {
		return nil, err
	}

	var outliers []int
	for i, value := range data {
		zScore := (value - mean) / stddev
		if zScore > threshold || zScore < -threshold {
			outliers = append(outliers, i)
		}
	}
	return outliers, nil
}

func IQROutliers(data []float64) ([]int, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	sort.Float64s(data)

	Q1, _, Q3, err := Quantiles(data)
	if err != nil {
		return nil, err
	}

	IQR := Q3 - Q1
	lowerOutlier := Q1 - 1.5*IQR
	upperOutlier := Q3 + 1.5*IQR

	var outliers []int
	for i, value := range data {
		if value < lowerOutlier || value > upperOutlier {
			outliers = append(outliers, i)
		}
	}

	return outliers, nil
}

func BoxPlotData(data []float64) (Q1, Q3, IQR, lowerOutlier, upperOutlier float64, err error) {
	if len(data) == 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("data cannot be empty")
	}

	Q1, _, Q3, err = Quantiles(data)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	IQR = Q3 - Q1
	lowerOutlier = Q1 - 1.5*IQR
	upperOutlier = Q3 + 1.5*IQR
	return Q1, Q3, IQR, lowerOutlier, upperOutlier, nil
}

func FormatBoxPlotForChartJS(Q1, Q3, lowerOutlier, upperOutlier float64) (labels []string, values []float64) {
	labels = []string{"Q1", "Q3", "Lower Outlier", "Upper Outlier"}
	values = []float64{Q1, Q3, lowerOutlier, upperOutlier}
	return labels, values
}

func Quantiles(data []float64) (Q1, Q2, Q3 float64, err error) {
	if len(data) == 0 {
		return 0, 0, 0, fmt.Errorf("data must not be empty")
	}

	sort.Float64s(data)

	median, err := descriptives.Median(data)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error calculating median of data: %v", data)
	}

	mid := len(data) / 2
	var lowerHalf, upperHalf []float64

	if len(data)%2 == 0 {
		lowerHalf = data[:mid]
		upperHalf = data[mid:]
	} else {
		lowerHalf = data[:mid]
		upperHalf = data[mid+1:]
	}

	Q1, err = descriptives.Median(lowerHalf)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error calculating Q1 of data: %v", data)
	}

	Q3, err = descriptives.Median(upperHalf)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error calculating Q3 of data: %v", data)
	}

	return Q1, median, Q3, nil
}
