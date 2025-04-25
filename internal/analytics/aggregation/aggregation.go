package aggregation

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

type GroupedResult map[string][]float64

// GroupBy groups rows based on a column and returns mapped values.
func GroupBy(data [][]string, keyCol int, valCol int) (GroupedResult, error) {
	result := make(GroupedResult)

	for i, row := range data {
		if len(row) <= keyCol || len(row) <= valCol {
			return nil, fmt.Errorf("row %d out of range for keyCol=%d or valCol=%d", i, keyCol, valCol)
		}

		key := row[keyCol]
		valStr := row[valCol]
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing float at row: %d: %v", i, err)
		}

		result[key] = append(result[key], val)
	}
	return result, nil
}

// GroupedSum returns the sum of values grouped by key.
func GroupedSum(groups GroupedResult) map[string]float64 {
	sums := make(map[string]float64)
	for key, values := range groups {
		total := 0.0
		for _, v := range values {
			total += v
		}
		sums[key] = total
	}
	return sums
}

// GroupedMean returns the average of values grouped by key.
func GroupedMean(groups GroupedResult) map[string]float64 {
	means := make(map[string]float64)
	for key, values := range groups {
		if len(values) == 0 {
			means[key] = 0
			continue
		}
		total := 0.0
		for _, v := range values {
			total += v
		}
		means[key] = total / float64(len(values))
	}
	return means
}

// GroupedCount returns the count of items per group.
func GroupedCount(groups GroupedResult) map[string]int {
	counts := make(map[string]int)
	for key, values := range groups {
		counts[key] = len(values)
	}
	return counts
}

// GroupedMin returns the minimum value in each group.
func GroupedMin(groups GroupedResult) map[string]float64 {
	mins := make(map[string]float64)
	for key, values := range groups {
		if len(values) == 0 {
			continue
		}
		min := values[0]
		for _, v := range values {
			if v < min {
				min = v
			}
		}
		mins[key] = min
	}
	return mins
}

// GroupedMax returns the maximum value in each group.
func GroupedMax(groups GroupedResult) map[string]float64 {
	maxs := make(map[string]float64)
	for key, values := range groups {
		if len(values) == 0 {
			continue
		}
		max := values[0]
		for _, v := range values {
			if v > max {
				max = v
			}
		}
		maxs[key] = max
	}
	return maxs
}

func GroupedMedian(groups GroupedResult) map[string]float64 {
	medians := make(map[string]float64)
	for key, values := range groups {
		if len(values) == 0 {
			medians[key] = 0
			continue
		}
		sorted := make([]float64, len(values))
		copy(sorted, values)
		sort.Float64s(sorted)
		n := len(sorted)
		if n%2 == 0 {
			medians[key] = (sorted[n/2-1] + sorted[n/2]) / 2
		} else {
			medians[key] = sorted[n/2]
		}
	}
	return medians
}

func GroupedStdDev(groups GroupedResult) map[string]float64 {
	stddevs := make(map[string]float64)
	for key, values := range groups {
		if len(values) == 0 {
			stddevs[key] = 0
			continue
		}
		mean := 0.0
		for _, v := range values {
			mean += v
		}
		mean /= float64(len(values))

		var sumSqDiff float64
		for _, v := range values {
			diff := v - mean
			sumSqDiff += diff * diff
		}
		stddevs[key] = math.Sqrt(sumSqDiff / float64(len(values)-1))
	}
	return stddevs
}

type PivotTable map[string]map[string]float64

func Pivot(data [][]string, rowKeyCol, colKeyCol, valCol int, aggFunc func([]float64) float64) (PivotTable, error) {
	table := make(PivotTable)
	temp := make(map[string]map[string][]float64)

	for i, row := range data {
		if len(row) <= valCol || len(row) <= rowKeyCol || len(row) <= colKeyCol {
			return nil, fmt.Errorf("row %d out of range", i)
		}

		rowKey := row[rowKeyCol]
		colKey := row[colKeyCol]
		valStr := row[valCol]

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing float at row %d: %v", i, err)
		}

		if temp[rowKey] == nil {
			temp[rowKey] = make(map[string][]float64)
		}
		temp[rowKey][colKey] = append(temp[rowKey][colKey], val)
	}

	for rowKey, colMap := range temp {
		table[rowKey] = make(map[string]float64)
		for colKey, values := range colMap {
			table[rowKey][colKey] = aggFunc(values)
		}
	}

	return table, nil
}

func PivotSum(data [][]string, rowKeyCol, colKeyCol, valCol int) (PivotTable, error) {
	return Pivot(data, rowKeyCol, colKeyCol, valCol, func(vals []float64) float64 {
		sum := 0.0
		for _, v := range vals {
			sum += v
		}
		return sum
	})
}

func PivotMean(data [][]string, rowKeyCol, colKeyCol, valCol int) (PivotTable, error) {
	return Pivot(data, rowKeyCol, colKeyCol, valCol, func(vals []float64) float64 {
		sum := 0.0
		for _, v := range vals {
			sum += v
		}
		return sum / float64(len(vals))
	})
}

func PivotMin(data [][]string, rowKeyCol, colKeyCol, valCol int) (PivotTable, error) {
	return Pivot(data, rowKeyCol, colKeyCol, valCol, func(vals []float64) float64 {
		min := vals[0]
		for _, v := range vals {
			if v < min {
				min = v
			}
		}
		return min
	})
}

func PivotMax(data [][]string, rowKeyCol, colKeyCol, valCol int) (PivotTable, error) {
	return Pivot(data, rowKeyCol, colKeyCol, valCol, func(vals []float64) float64 {
		max := vals[0]
		for _, v := range vals {
			if v > max {
				max = v
			}
		}
		return max
	})
}

func PivotCount(data [][]string, rowKeyCol, colKeyCol, valCol int) (PivotTable, error) {
	return Pivot(data, rowKeyCol, colKeyCol, valCol, func(vals []float64) float64 {
		return float64(len(vals))
	})
}

func PivotMedian(data [][]string, rowKeyCol, colKeyCol, valCol int) (PivotTable, error) {
	return Pivot(data, rowKeyCol, colKeyCol, valCol, func(vals []float64) float64 {
		sorted := append([]float64{}, vals...)
		sort.Float64s(sorted)
		mid := len(sorted) / 2
		if len(sorted)%2 == 0 {
			return (sorted[mid-1] + sorted[mid]) / 2
		}
		return sorted[mid]
	})
}

func PivotStdDev(data [][]string, rowKeyCol, colKeyCol, valCol int) (PivotTable, error) {
	return Pivot(data, rowKeyCol, colKeyCol, valCol, func(vals []float64) float64 {
		if len(vals) <= 1 {
			return 0
		}
		mean := 0.0
		for _, v := range vals {
			mean += v
		}
		mean /= float64(len(vals))
		variance := 0.0
		for _, v := range vals {
			diff := v - mean
			variance += diff * diff
		}
		variance /= float64(len(vals) - 1)
		return math.Sqrt(variance)
	})
}
