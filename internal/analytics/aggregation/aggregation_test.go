package aggregation_test

import (
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/analytics/aggregation"
	"github.com/stretchr/testify/assert"
)

var sampleData = [][]string{
	{"A", "X", "10"},
	{"A", "Y", "20"},
	{"B", "X", "30"},
	{"B", "Y", "40"},
	{"A", "X", "15"},
}

func TestGroupBy(t *testing.T) {
	grouped, err := aggregation.GroupBy(sampleData, 0, 2)
	assert.NoError(t, err)
	assert.Len(t, grouped, 2)
	assert.ElementsMatch(t, grouped["A"], []float64{10, 20, 15})
	assert.ElementsMatch(t, grouped["B"], []float64{30, 40})
}

func TestGroupedSum(t *testing.T) {
	grouped, _ := aggregation.GroupBy(sampleData, 0, 2)
	result := aggregation.GroupedSum(grouped)
	assert.Equal(t, 45.0, result["A"])
	assert.Equal(t, 70.0, result["B"])
}

func TestGroupedMean(t *testing.T) {
	grouped, _ := aggregation.GroupBy(sampleData, 0, 2)
	result := aggregation.GroupedMean(grouped)
	assert.InDelta(t, 15.0, result["A"], 0.01)
	assert.InDelta(t, 35.0, result["B"], 0.01)
}

func TestGroupedMedian(t *testing.T) {
	grouped, _ := aggregation.GroupBy(sampleData, 0, 2)
	result := aggregation.GroupedMedian(grouped)
	assert.Equal(t, 15.0, result["A"])
	assert.Equal(t, 35.0, result["B"])
}

func TestGroupedCount(t *testing.T) {
	groups := aggregation.GroupedResult{
		"A": {10, 20, 30},
		"B": {40},
		"C": {},
	}
	counts := aggregation.GroupedCount(groups)

	assert.Equal(t, 3, counts["A"])
	assert.Equal(t, 1, counts["B"])
	assert.Equal(t, 0, counts["C"])
}

func TestGroupedMin(t *testing.T) {
	groups := aggregation.GroupedResult{
		"A": {10, 20, 5},
		"B": {40, 30, 50},
		"C": {},
	}
	mins := aggregation.GroupedMin(groups)

	assert.Equal(t, 5.0, mins["A"])
	assert.Equal(t, 30.0, mins["B"])
	_, exists := mins["C"]
	assert.False(t, exists, "empty groups should not have a min")
}

func TestGroupedMax(t *testing.T) {
	groups := aggregation.GroupedResult{
		"A": {10, 20, 5},
		"B": {40, 30, 50},
		"C": {},
	}
	maxs := aggregation.GroupedMax(groups)

	assert.Equal(t, 20.0, maxs["A"])
	assert.Equal(t, 50.0, maxs["B"])
	_, exists := maxs["C"]
	assert.False(t, exists, "empty groups should not have a max")
}

func TestGroupedStdDev(t *testing.T) {
	grouped, _ := aggregation.GroupBy(sampleData, 0, 2)
	result := aggregation.GroupedStdDev(grouped)
	assert.InDelta(t, 5.0, result["A"], 0.01)
	assert.InDelta(t, 7.07, result["B"], 0.01)
}

func TestPivotSum(t *testing.T) {
	table, err := aggregation.PivotSum(sampleData, 0, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 25.0, table["A"]["X"])
	assert.Equal(t, 20.0, table["A"]["Y"])
	assert.Equal(t, 30.0, table["B"]["X"])
	assert.Equal(t, 40.0, table["B"]["Y"])
}

func TestPivotMean(t *testing.T) {
	table, err := aggregation.PivotMean(sampleData, 0, 1, 2)
	assert.NoError(t, err)
	assert.InDelta(t, 12.5, table["A"]["X"], 0.01)
	assert.InDelta(t, 20.0, table["A"]["Y"], 0.01)
	assert.InDelta(t, 30.0, table["B"]["X"], 0.01)
	assert.InDelta(t, 40.0, table["B"]["Y"], 0.01)
}

func TestPivotMin(t *testing.T) {
	table, err := aggregation.PivotMin(sampleData, 0, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 10.0, table["A"]["X"])
	assert.Equal(t, 20.0, table["A"]["Y"])
	assert.Equal(t, 30.0, table["B"]["X"])
	assert.Equal(t, 40.0, table["B"]["Y"])
}

func TestPivotMax(t *testing.T) {
	table, err := aggregation.PivotMax(sampleData, 0, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 15.0, table["A"]["X"])
	assert.Equal(t, 20.0, table["A"]["Y"])
	assert.Equal(t, 30.0, table["B"]["X"])
	assert.Equal(t, 40.0, table["B"]["Y"])
}

func TestPivotCount(t *testing.T) {
	table, err := aggregation.PivotCount(sampleData, 0, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2.0, table["A"]["X"])
	assert.Equal(t, 1.0, table["A"]["Y"])
	assert.Equal(t, 1.0, table["B"]["X"])
	assert.Equal(t, 1.0, table["B"]["Y"])
}

func TestPivotMedian(t *testing.T) {
	table, err := aggregation.PivotMedian(sampleData, 0, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 12.5, table["A"]["X"])
	assert.Equal(t, 20.0, table["A"]["Y"])
	assert.Equal(t, 30.0, table["B"]["X"])
	assert.Equal(t, 40.0, table["B"]["Y"])
}

func TestPivotStdDev(t *testing.T) {
	table, err := aggregation.PivotStdDev(sampleData, 0, 1, 2)
	assert.NoError(t, err)
	assert.InDelta(t, 3.5355, table["A"]["X"], 0.001)
	assert.Equal(t, 0.0, table["A"]["Y"]) // only one value
	assert.Equal(t, 0.0, table["B"]["X"]) // only one value
	assert.Equal(t, 0.0, table["B"]["Y"]) // only one value
}
