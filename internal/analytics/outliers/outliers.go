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

	// Debugging print statements
	fmt.Printf("Mean: %v, StdDev: %v\n", mean, stddev)

	var outliers []int
	for i, value := range data {
		zScore := (value - mean) / stddev
		// Debugging print statement
		fmt.Printf("Index: %d, Value: %v, Z-Score: %v\n", i, value, zScore)
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

	// Sort data to ensure it's in the correct order for quantiles calculation
	sort.Float64s(data)

	// Get quantiles
	Q1, _, Q3, err := Quantiles(data)
	if err != nil {
		return nil, err
	}

	// Calculate IQR (Interquartile Range)
	IQR := Q3 - Q1
	lowerOutlier := Q1 - 1.5*IQR // Changed to 1.5 multiplier
	upperOutlier := Q3 + 1.5*IQR // Changed to 1.5 multiplier

	// Debugging print statements
	fmt.Printf("Q1: %v, Q3: %v, IQR: %v, Lower Outlier: %v, Upper Outlier: %v\n", Q1, Q3, IQR, lowerOutlier, upperOutlier)

	var outliers []int
	for i, value := range data {
		// Debugging the value check
		fmt.Printf("Checking value: %v at index: %d\n", value, i)
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

	// Calculate the overall median (Q2)
	median, err := descriptives.Median(data)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error calculating median of data: %v", data)
	}

	// Split data into halves based on the dataset's length
	mid := len(data) / 2
	var lowerHalf, upperHalf []float64

	if len(data)%2 == 0 {
		// Even number of elements: No exclusion of median, split evenly
		lowerHalf = data[:mid]
		upperHalf = data[mid:]
	} else {
		// Odd number of elements: Exclude the median value
		lowerHalf = data[:mid]
		upperHalf = data[mid+1:]
	}

	// Calculate Q1 (median of the lower half) and Q3 (median of the upper half)
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
