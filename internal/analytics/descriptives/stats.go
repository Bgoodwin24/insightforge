package descriptives

import (
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

func Mode(data []float64) ([]float64, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	counts := make(map[float64]int)
	maxCount := 0
	for _, num := range data {
		counts[num]++
		if counts[num] > maxCount {
			maxCount = counts[num]
		}
	}

	var modes []float64
	for num, count := range counts {
		if count == maxCount {
			modes = append(modes, num)
		}
	}

	return modes, nil
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
		return 0, fmt.Errorf("data is empty")
	}

	sumSquaredDiff := 0.0
	mean, err := Mean(data)
	if err != nil {
		return 0.0, err
	}

	for _, val := range data {
		diff := val - mean
		sumSquaredDiff += diff * diff
	}
	return sumSquaredDiff / float64(len(data)-1), nil
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
