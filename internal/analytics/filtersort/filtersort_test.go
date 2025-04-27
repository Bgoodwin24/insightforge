package filtersort_test

import (
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/analytics/filtersort"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFirst(t *testing.T) {
	params := map[string][]string{
		"key1": {"value1"},
		"key2": {},
	}
	assert.Equal(t, "value1", filtersort.GetFirst(params, "key1"))
	assert.Equal(t, "", filtersort.GetFirst(params, "key2"))
	assert.Equal(t, "", filtersort.GetFirst(params, "missing"))
}

func TestParseFilterSort(t *testing.T) {
	params := map[string][]string{
		"sort_by":    {"age"},
		"order":      {"desc"},
		"filter_col": {"city"},
		"filter_op":  {"contains"},
		"filter_val": {"York"},
	}

	filters, sortOpt, err := filtersort.ParseFilterSort(params)
	require.NoError(t, err)
	require.Len(t, filters, 1)
	require.NotNil(t, sortOpt)

	assert.Equal(t, "city", filters[0].Column)
	assert.Equal(t, "contains", filters[0].Op)
	assert.Equal(t, "York", filters[0].Value)

	assert.Equal(t, "age", sortOpt.Column)
	assert.Equal(t, "desc", sortOpt.Order)
}

func TestApplySort(t *testing.T) {
	data := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
		{"Charlie", "35"},
	}
	headers := []string{"name", "age"}

	sortOpt := &filtersort.SortOption{
		Column: "age",
		Order:  "asc",
	}

	sorted, err := filtersort.ApplySort(data, headers, sortOpt)
	require.NoError(t, err)

	assert.Equal(t, "25", sorted[0][1])
	assert.Equal(t, "30", sorted[1][1])
	assert.Equal(t, "35", sorted[2][1])
}

func TestApplyFilterSort(t *testing.T) {
	data := [][]string{
		{"Alice", "New York", "30"},
		{"Bob", "Los Angeles", "25"},
		{"Charlie", "New York", "35"},
	}
	headers := []string{"name", "city", "age"}

	filters := []filtersort.FilterOption{
		{Column: "city", Op: "contains", Value: "York"},
	}
	sortOpt := &filtersort.SortOption{
		Column: "age",
		Order:  "asc",
	}

	filteredSorted, err := filtersort.ApplyFilterSort(data, headers, filters, sortOpt)
	require.NoError(t, err)

	require.Len(t, filteredSorted, 2)
	assert.Equal(t, "Alice", filteredSorted[0][0])   // Alice is 30
	assert.Equal(t, "Charlie", filteredSorted[1][0]) // Charlie is 35
}
