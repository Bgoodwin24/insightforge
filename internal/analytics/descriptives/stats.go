package descriptives

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

func Mean(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data is empty")
	}
	count := 0.0
	amount := len(data)
	for i := 0; i < amount; i++ {
		count += data[i]
	}

	return count / float64(amount), nil
}

func Median(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data is empty")
	}

	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	mid := len(sortedData) / 2
	if len(sortedData)%2 == 0 {
		return (sortedData[mid-1] + sortedData[mid]) / 2, nil
	}
	return sortedData[mid], nil
}

func Mode(data []string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("no data provided")
	}

	freq := make(map[string]int)
	maxCount := 0
	var mode string

	for _, v := range data {
		freq[v]++
		if freq[v] > maxCount {
			maxCount = freq[v]
			mode = v
		}
	}

	return mode, nil
}

func StdDev(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data is empty")
	}

	variance, err := Variance(data)
	if err != nil {
		return 0, err
	}
	return math.Sqrt(variance), nil
}

func Variance(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, errors.New("no data provided")
	}

	var sum, mean, variance float64
	n := float64(len(data))

	for _, value := range data {
		sum += value
	}
	mean = sum / n

	for _, value := range data {
		variance += (value - mean) * (value - mean)
	}
	variance /= (n - 1)

	return variance, nil
}

func Min(data []float64) (float64, error) {
	min := math.Inf(1)
	if len(data) == 0 {
		return min, fmt.Errorf("data is empty")
	}

	for _, val := range data {
		if val < min {
			min = val
		}
	}
	return min, nil
}

func Max(data []float64) (float64, error) {
	if len(data) == 0 {
		return math.Inf(-1), fmt.Errorf("data is empty")
	}

	max := data[0]
	for _, num := range data {
		if num > max {
			max = num
		}
	}

	return max, nil
}

func Range(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data is empty")
	}

	min := data[0]
	max := data[0]

	for _, num := range data {
		min = math.Min(min, num)
		max = math.Max(max, num)
	}
	return max - min, nil
}

func Sum(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data is empty")
	}

	sum := 0.0
	for _, num := range data {
		sum += num
	}
	return sum, nil
}

func Count(data []float64) int {
	return len(data)
}
