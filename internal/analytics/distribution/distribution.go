package distribution

import (
	"fmt"
	"math"
	"sort"
)

// Histogram computes the counts of values within evenly spaced bins.
func Histogram(data []float64, numBins int) (binEdges []float64, binCounts []int, err error) {
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("no data points provided")
	}
	if numBins <= 0 {
		return nil, nil, fmt.Errorf("number of bins must be positive")
	}

	// Find min and max of data
	minVal := data[0]
	maxVal := data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Handle special case: all data points identical
	if minVal == maxVal {
		minVal -= 0.5
		maxVal += 0.5
	}

	binWidth := (maxVal - minVal) / float64(numBins)
	binEdges = make([]float64, numBins+1)
	for i := 0; i <= numBins; i++ {
		binEdges[i] = minVal + binWidth*float64(i)
	}

	binCounts = make([]int, numBins)
	for _, v := range data {
		idx := int((v - minVal) / binWidth)
		// Edge case: max value falls exactly on upper edge
		if idx == numBins {
			idx--
		}
		binCounts[idx]++
	}
	return binEdges, binCounts, nil
}

// KDEApproximate estimates a smooth density curve over the data using a simple Gaussian kernel.
func KDEApproximate(data []float64, numPoints int, bandwidth float64) (xs []float64, ys []float64, err error) {
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("no data points provided")
	}
	if numPoints <= 0 {
		return nil, nil, fmt.Errorf("number of output points must be positive")
	}
	if bandwidth <= 0 {
		return nil, nil, fmt.Errorf("bandwidth must be positive")
	}

	sort.Float64s(data)
	minVal := data[0]
	maxVal := data[len(data)-1]

	// Generate evenly spaced xs across the data
	xs = make([]float64, numPoints)
	step := (maxVal - minVal) / float64(numPoints-1)
	for i := range xs {
		xs[i] = minVal + float64(i)*step
	}

	ys = make([]float64, numPoints)
	n := float64(len(data))

	for i, x := range xs {
		var sum float64
		for _, xi := range data {
			u := (x - xi) / bandwidth
			sum += math.Exp(-0.5 * u * u)
		}
		ys[i] = (1.0 / (n * bandwidth * math.Sqrt(2*math.Pi))) * sum
	}
	return xs, ys, nil
}

// FormatHistogramForChartJS formats histogram output into Chart.js-friendly labels and counts.
func FormatHistogramForChartJS(binEdges []float64, binCounts []int) (labels []string, counts []int) {
	n := len(binCounts)
	labels = make([]string, n)
	for i := 0; i < n; i++ {
		labels[i] = fmt.Sprintf("[%.2f, %.2f]", binEdges[i], binEdges[i+1])
	}
	return labels, binCounts
}

// FormatKDEForChartJS formats KDE output into Chart.js-friendly labels and densities.
func FormatKDEForChartJS(xs []float64, ys []float64) (labels []string, densities []float64) {
	n := len(xs)
	labels = make([]string, n)
	for i := 0; i < n; i++ {
		labels[i] = fmt.Sprintf("%.2f", xs[i])
	}
	return labels, ys
}
