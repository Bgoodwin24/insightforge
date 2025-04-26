package correlation

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

type CorrelationResult struct {
	Matrix [][]float64 `json:"matrix"`
	Labels []string    `json:"labels"`
}

// PearsonCorrelation computes the Pearson correlation coefficient between two float slices.
func PearsonCorrelation(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("input slices must be the same length")
	}
	n := len(x)
	if n == 0 {
		return 0, fmt.Errorf("input slices cannot be empty")
	}

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	num := float64(n)*sumXY - sumX*sumY
	den := math.Sqrt((float64(n)*sumX2 - sumX*sumX) * (float64(n)*sumY2 - sumY*sumY))
	if den == 0 {
		return 0, fmt.Errorf("division by zero in Pearson calculation")
	}

	return num / den, nil
}

// SpearmanCorrelation computes the Spearman rank correlation coefficient between two float slices.
func SpearmanCorrelation(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("input slices must be the same length")
	}
	rankX, err := Rank(x)
	if err != nil {
		return 0, err
	}
	rankY, err := Rank(y)
	if err != nil {
		return 0, err
	}
	return PearsonCorrelation(rankX, rankY)
}

// CorrelationMatrix returns an N x N matrix of correlation coefficients for the selected numeric columns.
// Method should be "pearson" or "spearman".
func CorrelationMatrix(data [][]string, colIndices []int, method string) ([][]float64, error) {
	cols, err := ExtractFloatColumns(data, colIndices)
	if err != nil {
		return nil, err
	}

	n := len(cols)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
		for j := 0; j <= i; j++ {
			var corr float64
			var err error
			switch method {
			case "pearson":
				corr, err = PearsonCorrelation(cols[i], cols[j])
			case "spearman":
				corr, err = SpearmanCorrelation(cols[i], cols[j])
			default:
				return nil, fmt.Errorf("unknown correlation method: %s", method)
			}
			if err != nil {
				return nil, err
			}
			matrix[i][j] = corr
			matrix[j][i] = corr // symmetric
		}
	}
	return matrix, nil
}

// ExtractFloatColumns parses selected columns as float64 slices from a 2D string dataset.
func ExtractFloatColumns(data [][]string, colIndices []int) ([][]float64, error) {
	result := make([][]float64, len(colIndices))
	for i := range result {
		result[i] = make([]float64, 0, len(data))
	}
	for _, row := range data {
		for j, colIdx := range colIndices {
			if colIdx >= len(row) {
				return nil, fmt.Errorf("column index %d out of range", colIdx)
			}
			val, err := strconv.ParseFloat(row[colIdx], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float at row: %v", err)
			}
			result[j] = append(result[j], val)
		}
	}
	return result, nil
}

// Rank converts a slice of float64s into ranks (handling ties with average rank).
func Rank(data []float64) ([]float64, error) {
	n := len(data)
	if n == 0 {
		return nil, fmt.Errorf("empty data for ranking")
	}

	type valIdx struct {
		val float64
		idx int
	}
	sorted := make([]valIdx, n)
	for i, v := range data {
		sorted[i] = valIdx{val: v, idx: i}
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].val < sorted[j].val
	})

	ranks := make([]float64, n)
	i := 0
	for i < n {
		j := i + 1
		for j < n && sorted[j].val == sorted[i].val {
			j++
		}
		rank := float64(i+j-1) / 2.0
		for k := i; k < j; k++ {
			ranks[sorted[k].idx] = rank
		}
		i = j
	}
	return ranks, nil
}

// GenerateCorrelationLabels creates a header row and column labels for a correlation matrix.
func GenerateCorrelationLabels(colIndices []int, headers []string) [][]string {
	n := len(colIndices)
	result := make([][]string, n+1)

	// Header row
	headerRow := make([]string, n+1)
	headerRow[0] = ""
	for i, idx := range colIndices {
		headerRow[i+1] = headers[idx]
	}
	result[0] = headerRow

	// Data rows with row labels
	for i, idx := range colIndices {
		row := make([]string, n+1)
		row[0] = headers[idx]
		for j := 1; j <= n; j++ {
			row[j] = "" // filled later by visualization
		}
		result[i+1] = row
	}

	return result
}

func FormatCorrelationResult(matrix [][]float64, colIndices []int, headers []string) CorrelationResult {
	labels := make([]string, len(colIndices))
	for i, idx := range colIndices {
		labels[i] = headers[idx]
	}
	return CorrelationResult{
		Matrix: matrix,
		Labels: labels,
	}
}
