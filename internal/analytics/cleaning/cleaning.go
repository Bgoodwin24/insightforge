package cleaning

import (
	"fmt"
	"math"
	"strconv"
)

// Missing Value Handling
// Drops any rows that contain empty string cells
func DropRowsWithMissing(data [][]string, columns []string) [][]string {
	if len(data) == 0 {
		return data
	}

	// Map column names to their index
	header := data[0]
	colIndexes := make([]int, 0, len(columns))
	for _, col := range columns {
		found := false
		for i, name := range header {
			if name == col {
				colIndexes = append(colIndexes, i)
				found = true
				break
			}
		}
		if !found {
			// If a column isn't found in header, skip or return empty
			return [][]string{header} // return only header if column doesn't exist
		}
	}

	// Build result, always keep header
	cleaned := [][]string{header}

	for _, row := range data[1:] {
		missing := false
		for _, idx := range colIndexes {
			if row[idx] == "" {
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
func NormalizeColumn(data [][]string, col int) ([]float64, error) {
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

	result := make([]float64, len(values))
	for i, v := range values {
		if max == min {
			result[i] = 0
		} else {
			norm := (v - min) / (max - min)
			result[i] = math.Round(norm*1e6) / 1e6
		}
	}

	return result, nil
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
