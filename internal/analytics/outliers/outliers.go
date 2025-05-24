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
	if len(outliers) == 0 {
		return []int{}, nil
	}
	return outliers, nil
}

func IQROutliers(data []float64) ([]int, float64, float64, error) {
	if len(data) == 0 {
		return nil, 0, 0, fmt.Errorf("data cannot be empty")
	}

	Q1, _, Q3, err := Quantiles(data)
	if err != nil {
		return nil, 0, 0, err
	}

	IQR := Q3 - Q1
	lowerBound := Q1 - 1.5*IQR
	upperBound := Q3 + 1.5*IQR

	var indices []int
	for i, value := range data {
		if value < lowerBound || value > upperBound {
			indices = append(indices, i)
		}
	}

	return indices, lowerBound, upperBound, nil
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

func Quantiles(data []float64) (float64, float64, float64, error) {
	if len(data) == 0 {
		return 0, 0, 0, fmt.Errorf("data cannot be empty")
	}

	sort.Float64s(data)

	getPercentile := func(p float64) float64 {
		index := p * float64(len(data)-1)
		lower := int(index)
		upper := lower + 1
		if upper >= len(data) {
			return data[lower]
		}
		weight := index - float64(lower)
		return data[lower]*(1-weight) + data[upper]*weight
	}

	Q1 := getPercentile(0.25)
	Q2 := getPercentile(0.50)
	Q3 := getPercentile(0.75)

	return Q1, Q2, Q3, nil
}
