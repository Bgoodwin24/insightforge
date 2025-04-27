package filtersort

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type FilterOption struct {
	Column string
	Op     string
	Value  string
}

type SortOption struct {
	Column string
	Order  string
}

func GetFirst(params map[string][]string, key string) string {
	values, ok := params[key]
	if !ok || len(values) == 0 {
		return ""
	}
	return values[0]
}

func ParseFilterSort(queryParams map[string][]string) ([]FilterOption, *SortOption, error) {
	var filters []FilterOption
	var sortOption *SortOption

	sortBy := GetFirst(queryParams, "sort_by")
	order := GetFirst(queryParams, "order")
	if sortBy != "" {
		if order == "" {
			order = "asc"
		}
		order = strings.ToLower(order)
		if order != "asc" && order != "desc" {
			return nil, nil, fmt.Errorf("invalid sort order: %s", order)
		}
		sortOption = &SortOption{
			Column: sortBy,
			Order:  order,
		}
	}

	filterCol := GetFirst(queryParams, "filter_col")
	filterOp := GetFirst(queryParams, "filter_op")
	filterVal := GetFirst(queryParams, "filter_val")
	if filterCol != "" && filterOp != "" && filterVal != "" {
		filters = append(filters, FilterOption{
			Column: filterCol,
			Op:     strings.ToLower(filterOp),
			Value:  filterVal,
		})
	}
	return filters, sortOption, nil
}

func ApplySort(data [][]string, headers []string, sortOption *SortOption) ([][]string, error) {
	if sortOption == nil {
		return data, nil
	}

	colIdx := -1
	for i, h := range headers {
		if h == sortOption.Column {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return nil, fmt.Errorf("sort column %s not found", sortOption.Column)
	}

	sorted := make([][]string, len(data))
	copy(sorted, data)

	sort.SliceStable(sorted, func(i, j int) bool {
		valI := sorted[i][colIdx]
		valJ := sorted[j][colIdx]

		// Try numeric comparison first
		numI, errI := strconv.ParseFloat(valI, 64)
		numJ, errJ := strconv.ParseFloat(valJ, 64)
		if errI == nil && errJ == nil {
			if sortOption.Order == "asc" {
				return numI < numJ
			}
			return numI > numJ
		}

		// Fall back to string comparison
		if sortOption.Order == "asc" {
			return strings.Compare(valI, valJ) < 0
		}
		return strings.Compare(valI, valJ) > 0
	})

	return sorted, nil
}

func ApplyFilterSort(data [][]string, headers []string, filters []FilterOption, sortOption *SortOption) ([][]string, error) {
	filtered := data

	// Apply each filter
	for _, f := range filters {
		colIdx := -1
		for i, h := range headers {
			if h == f.Column {
				colIdx = i
				break
			}
		}
		if colIdx == -1 {
			return nil, fmt.Errorf("filter column %s not found", f.Column)
		}

		var temp [][]string
		for _, row := range filtered {
			val := row[colIdx]
			match := false

			switch f.Op {
			case "eq":
				match = val == f.Value
			case "contains":
				match = strings.Contains(strings.ToLower(val), strings.ToLower(f.Value))
			case "gt", "lt", "ge", "le":
				numVal, err1 := strconv.ParseFloat(val, 64)
				numFilter, err2 := strconv.ParseFloat(f.Value, 64)
				if err1 == nil && err2 == nil {
					switch f.Op {
					case "gt":
						match = numVal > numFilter
					case "lt":
						match = numVal < numFilter
					case "ge":
						match = numVal >= numFilter
					case "le":
						match = numVal <= numFilter
					}
				}
			default:
				return nil, fmt.Errorf("unsupported filter operation: %s", f.Op)
			}

			if match {
				temp = append(temp, row)
			}
		}
		filtered = temp
	}

	// Apply sorting after filtering
	sorted, err := ApplySort(filtered, headers, sortOption)
	if err != nil {
		return nil, err
	}

	return sorted, nil
}
