package cleaning

import (
	"fmt"
	"math"
	"strconv"
)

// Missing Value Handling
// Drops any rows that contain empty string cells
func DropRowsWithMissing(data [][]string) [][]string {
	var cleaned [][]string
	for _, row := range data {
		missing := false
		for _, cell := range row {
			if cell == "" {
				missing = true
				break
			}
		}
		if !missing {
			cleaned = append(cleaned, row)
		}
	}
	return cleaned
}

// Replaces empty cells with default value
func FillMissingWith(data [][]string, defaultValue string) [][]string {
	cleaned := make([][]string, len(data))
	for i, row := range data {
		cleaned[i] = make([]string, len(row))
		for j, cell := range row {
			if cell == "" {
				cleaned[i][j] = defaultValue
			} else {
				cleaned[i][j] = cell
			}
		}
	}
	return cleaned
}

// Type Conversion Helpers
// Extracts a specific column as float64s, skipping rows with non-convertable data
func ToFloatSlice(data [][]string, col int) ([]float64, error) {
	var result []float64
	for _, row := range data {
		if len(row) <= col {
			continue
		}
		val, err := strconv.ParseFloat(row[col], 64)
		if err != nil {
			continue // skip non-numeric rows
		}
		result = append(result, val)
	}
	return result, nil
}

// Transformations
// Applies log(x) to a column of numeric data
func ApplyLogTransformation(data [][]string, col int) ([][]string, error) {
	transformed := make([][]string, len(data))
	for i, row := range data {
		if len(row) <= col {
			transformed[i] = row
			continue
		}
		val, err := strconv.ParseFloat(row[col], 64)
		if err != nil || val <= 0 {
			return nil, fmt.Errorf("invalid value for log transform")
		}
		newRow := make([]string, len(row))
		copy(newRow, row)
		newRow[col] = strconv.FormatFloat(math.Log(val), 'f', -1, 64)
		transformed[i] = newRow
	}
	return transformed, nil
}

// Applies min-max normalization to a column
func NormalizeColumn(data [][]string, col int) ([][]string, error) {
	values := make([]float64, len(data))

	for i, row := range data {
		if len(row) <= col {
			return nil, fmt.Errorf("row %d out of range", i)
		}
		v, err := strconv.ParseFloat(row[col], 64)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	normalized := make([][]string, len(data))
	for i := range data {
		normalized[i] = make([]string, len(data[i]))
		copy(normalized[i], data[i])
		if max == min {
			normalized[i][col] = "0" // Avoid divide by zero
		} else {
			norm := (values[i] - min) / (max - min)
			normalized[i][col] = fmt.Sprintf("%f", norm)
		}
	}

	return normalized, nil
}

// Applies z-score standardization to a column
func StandardizeColumn(data [][]string, col int) ([][]string, error) {
	values, err := ToFloatSlice(data, col)
	if err != nil || len(values) == 0 {
		return nil, fmt.Errorf("could not parse column for standardization")
	}
	var sum, sumSq float64
	for _, v := range values {
		sum += v
		sumSq += v * v
	}
	mean := sum / float64(len(values))
	variance := (sumSq / float64(len(values))) - (mean * mean)
	stdDev := math.Sqrt(variance)
	if stdDev == 0 {
		return nil, fmt.Errorf("standard deviation is zero")
	}
	standardized := make([][]string, len(data))
	for i, row := range data {
		newRow := make([]string, len(row))
		copy(newRow, row)
		val, err := strconv.ParseFloat(row[col], 64)
		if err == nil {
			z := (val - mean) / stdDev
			newRow[col] = strconv.FormatFloat(z, 'f', -1, 64)
		}
		standardized[i] = newRow
	}
	return standardized, nil
}

// Column Operations
// Removes columns at the given index
func DropColumns(data [][]string, col []int) [][]string {
	colSet := make(map[int]bool)
	for _, idx := range col {
		colSet[idx] = true
	}
	var cleaned [][]string
	for _, row := range data {
		var newRow []string
		for j, val := range row {
			if !colSet[j] {
				newRow = append(newRow, val)
			}
		}
		cleaned = append(cleaned, newRow)
	}
	return cleaned
}

// Returns a copy of the data with a new header row
func RenameColumns(data [][]string, newHeaders []string) ([][]string, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty dataset")
	}
	if len(newHeaders) != len(data[0]) {
		return nil, fmt.Errorf("header length mismatch")
	}
	out := make([][]string, len(data))
	out[0] = make([]string, len(newHeaders))
	copy(out[0], newHeaders)
	for i := 1; i < len(data); i++ {
		out[i] = data[i]
	}
	return out, nil
}
