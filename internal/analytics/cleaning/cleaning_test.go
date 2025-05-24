package cleaning_test

import (
	"strconv"
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/analytics/cleaning"
	"github.com/stretchr/testify/assert"
)

func TestDropRowsWithMissing(t *testing.T) {
	data := [][]string{
		{"name", "value"}, // header row
		{"a", "1"},
		{"b", ""},
		{"", "3"},
		{"c", "2"},
	}
	expected := [][]string{
		{"name", "value"},
		{"a", "1"},
		{"c", "2"},
	}
	columns := []string{"name", "value"}

	output := cleaning.DropRowsWithMissing(data, columns)
	assert.Equal(t, expected, output)
}

func TestFillMissingWith(t *testing.T) {
	input := [][]string{
		{"a", ""},
		{"", "2"},
	}
	expected := [][]string{
		{"a", "N/A"},
		{"N/A", "2"},
	}
	output := cleaning.FillMissingWith(input, "N/A")
	assert.Equal(t, expected, output)
}

func TestToFloatSlice(t *testing.T) {
	input := [][]string{
		{"a", "1.5"},
		{"b", "2"},
		{"c", "invalid"},
	}
	expected := []float64{1.5, 2.0}
	output, err := cleaning.ToFloatSlice(input, 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

func TestApplyLogTransformation(t *testing.T) {
	input := [][]string{
		{"a", "1"},
		{"b", "2"},
		{"c", "3"},
	}
	output, err := cleaning.ApplyLogTransformation(input, 1)
	assert.NoError(t, err)
	assert.InDelta(t, 0.0, mustParseFloat(output[0][1]), 0.0001)
	assert.InDelta(t, 0.6931, mustParseFloat(output[1][1]), 0.0001)
	assert.InDelta(t, 1.0986, mustParseFloat(output[2][1]), 0.0001)
}

func TestNormalizeColumn(t *testing.T) {
	input := [][]string{
		{"a", "1"},
		{"b", "2"},
		{"c", "3"},
	}

	output, err := cleaning.NormalizeColumn(input, 1)
	assert.NoError(t, err)
	assert.Len(t, output, 3)

	assert.InDelta(t, 0.0, output[0], 0.0001)
	assert.InDelta(t, 0.5, output[1], 0.0001)
	assert.InDelta(t, 1.0, output[2], 0.0001)
}

func TestStandardizeColumn(t *testing.T) {
	input := [][]string{
		{"a", "1"},
		{"b", "2"},
		{"c", "3"},
	}
	output, err := cleaning.StandardizeColumn(input, 1)
	assert.NoError(t, err)
	assert.InDelta(t, -1.2247, mustParseFloat(output[0][1]), 0.0001)
	assert.InDelta(t, 0.0, mustParseFloat(output[1][1]), 0.0001)
	assert.InDelta(t, 1.2247, mustParseFloat(output[2][1]), 0.0001)
}

func TestDropColumns(t *testing.T) {
	input := [][]string{
		{"a", "1", "x"},
		{"b", "2", "y"},
	}
	expected := [][]string{
		{"a", "x"},
		{"b", "y"},
	}
	output := cleaning.DropColumns(input, []int{1})
	assert.Equal(t, expected, output)
}

func TestRenameColumns(t *testing.T) {
	input := [][]string{
		{"old1", "old2"},
		{"val1", "val2"},
	}
	newHeaders := []string{"new1", "new2"}
	expected := [][]string{
		{"new1", "new2"},
		{"val1", "val2"},
	}
	output, err := cleaning.RenameColumns(input, newHeaders)
	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

// Helper to parse float in tests
func mustParseFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("test float parse failed")
	}
	return val
}
