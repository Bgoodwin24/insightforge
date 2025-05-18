package handlers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Bgoodwin24/insightforge/internal/analytics/aggregation"
	"github.com/Bgoodwin24/insightforge/internal/analytics/cleaning"
	"github.com/Bgoodwin24/insightforge/internal/analytics/correlation"
	"github.com/Bgoodwin24/insightforge/internal/analytics/descriptives"
	"github.com/Bgoodwin24/insightforge/internal/analytics/distribution"
	"github.com/Bgoodwin24/insightforge/internal/analytics/outliers"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func convertToFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

type AnalyticsHandler struct {
	Service        *services.DatasetService
	DatasetService *services.DatasetService
}

const userIDKey = "user_id"

// GetUserFromContext returns the authenticated user's ID from Gin context
func GetUserFromContext(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(userIDKey)
	if !exists {
		return uuid.Nil, errors.New("user not found in context")
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID format in context")
	}

	return userID, nil
}

// Aggregation
func (h *AnalyticsHandler) GroupDatasetBy(ctx context.Context, datasetID, userID uuid.UUID, groupBy, column string) (map[string][]float64, error) {
	header, rows, err := h.Service.GetDatasetRows(ctx, datasetID, userID)
	if err != nil {
		return nil, err
	}

	// Find index of groupBy and column
	groupByIdx, columnIdx := -1, -1
	for i, col := range header {
		if col == groupBy {
			groupByIdx = i
		}
		if col == column {
			columnIdx = i
		}
	}
	if groupByIdx == -1 || columnIdx == -1 {
		return nil, fmt.Errorf("group_by or column not found in dataset")
	}

	grouped := make(map[string][]float64)
	for _, row := range rows {
		group := row[groupByIdx]
		val, ok := convertToFloat64(row[columnIdx])
		if !ok {
			continue
		}
		grouped[group] = append(grouped[group], val)
	}
	return grouped, nil
}

func (h *AnalyticsHandler) GroupedSumHandler(c *gin.Context) {
	datasetIDStr, groupBy, column := c.Query("dataset_id"), c.Query("group_by"), c.Query("column")

	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil || groupBy == "" || column == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid dataset_id, group_by, or column"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	grouped, err := h.GroupDatasetBy(c.Request.Context(), datasetID, userID, groupBy, column)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aggregation.GroupedSum(grouped))
}

func (h *AnalyticsHandler) GroupedMeanHandler(c *gin.Context) {
	datasetID, groupBy, column := c.Query("dataset_id"), c.Query("group_by"), c.Query("column")
	id, err := uuid.Parse(datasetID)
	if err != nil || groupBy == "" || column == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid dataset_id, group_by, or column"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grouped, err := h.GroupDatasetBy(c.Request.Context(), id, userID, groupBy, column)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aggregation.GroupedMean(grouped))
}

func (h *AnalyticsHandler) GroupedCountHandler(c *gin.Context) {
	datasetID, groupBy := c.Query("dataset_id"), c.Query("group_by")
	id, err := uuid.Parse(datasetID)
	if err != nil || groupBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid dataset_id or group_by"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	header, rows, err := h.Service.GetDatasetRows(c, id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the index of groupBy
	groupByIdx := -1
	for i, col := range header {
		if col == groupBy {
			groupByIdx = i
			break
		}
	}
	if groupByIdx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_by column not found in dataset"})
		return
	}

	grouped := make(map[string][]float64)
	for _, row := range rows {
		group := row[groupByIdx]
		grouped[group] = append(grouped[group], 1)
	}

	c.JSON(http.StatusOK, aggregation.GroupedCount(grouped))
}

func (h *AnalyticsHandler) GroupedMinHandler(c *gin.Context) {
	datasetID, groupBy, column := c.Query("dataset_id"), c.Query("group_by"), c.Query("column")
	id, err := uuid.Parse(datasetID)
	if err != nil || groupBy == "" || column == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid dataset_id, group_by, or column"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grouped, err := h.GroupDatasetBy(c.Request.Context(), id, userID, groupBy, column)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aggregation.GroupedMin(grouped))
}

func (h *AnalyticsHandler) GroupedMaxHandler(c *gin.Context) {
	datasetID, groupBy, column := c.Query("dataset_id"), c.Query("group_by"), c.Query("column")
	id, err := uuid.Parse(datasetID)
	if err != nil || groupBy == "" || column == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid dataset_id, group_by, or column"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grouped, err := h.GroupDatasetBy(c.Request.Context(), id, userID, groupBy, column)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aggregation.GroupedMax(grouped))
}

func (h *AnalyticsHandler) GroupedMedianHandler(c *gin.Context) {
	datasetID, groupBy, column := c.Query("dataset_id"), c.Query("group_by"), c.Query("column")
	id, err := uuid.Parse(datasetID)
	if err != nil || groupBy == "" || column == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid dataset_id, group_by, or column"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grouped, err := h.GroupDatasetBy(c.Request.Context(), id, userID, groupBy, column)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aggregation.GroupedMedian(grouped))
}

func (h *AnalyticsHandler) GroupedStdDevHandler(c *gin.Context) {
	datasetID, groupBy, column := c.Query("dataset_id"), c.Query("group_by"), c.Query("column")
	id, err := uuid.Parse(datasetID)
	if err != nil || groupBy == "" || column == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid dataset_id, group_by, or column"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grouped, err := h.GroupDatasetBy(c.Request.Context(), id, userID, groupBy, column)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aggregation.GroupedStdDev(grouped))
}

type PivotRequest struct {
	RowField    string                   `json:"row_field"`
	ColumnField string                   `json:"column_field"`
	ValueField  string                   `json:"value_field"`
	AggFunc     string                   `json:"agg_func"`
	Data        []map[string]interface{} `json:"data"`
}

func (h *AnalyticsHandler) PivotTableHandler(c *gin.Context) {
	datasetIDStr := c.Query("dataset_id")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dataset_id"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get dataset rows ([][]string)
	header, rows, err := h.Service.GetDatasetRows(c.Request.Context(), datasetID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse query parameters for pivot settings
	rowField := c.Query("row_field")
	colField := c.Query("column_field")
	valField := c.Query("value_field")
	aggFunc := c.Query("agg_func")
	if aggFunc == "" {
		// Fallback: detect from route if needed
		path := c.FullPath()
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			last := parts[len(parts)-1]
			if strings.HasPrefix(last, "pivot-") {
				aggFunc = strings.TrimPrefix(last, "pivot-")
			}
		}
	}

	if rowField == "" || colField == "" || (aggFunc != "count" && valField == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required pivot parameters"})
		return
	}

	// Convert rows [][]string with headers in rows[0] into []map[string]interface{}
	// for easier access by field name
	data := []map[string]interface{}{}
	for _, row := range rows {
		if len(row) != len(header) {
			continue
		}
		record := make(map[string]interface{})
		for i, col := range header {
			record[col] = row[i]
		}
		data = append(data, record)
	}

	// Convert data to [][]string expected by aggregation pivot functions
	pivotData := [][]string{}
	for i, record := range data {
		rowVal, ok1 := record[rowField]
		colVal, ok2 := record[colField]
		if !ok1 || !ok2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing row_field or column_field in data row %d", i)})
			return
		}
		rowStr := fmt.Sprintf("%v", rowVal)
		colStr := fmt.Sprintf("%v", colVal)

		valStr := "1" // Default for count
		if aggFunc != "count" {
			valVal, ok3 := record[valField]
			if !ok3 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing value_field in data row %d", i)})
				return
			}
			valStr = fmt.Sprintf("%v", valVal)
		}

		pivotData = append(pivotData, []string{rowStr, colStr, valStr})
	}

	// Run aggregation pivot function based on aggFunc
	var pivotTable aggregation.PivotTable
	switch strings.ToLower(aggFunc) {
	case "sum":
		pivotTable, err = aggregation.PivotSum(pivotData, 0, 1, 2)
	case "mean":
		pivotTable, err = aggregation.PivotMean(pivotData, 0, 1, 2)
	case "min":
		pivotTable, err = aggregation.PivotMin(pivotData, 0, 1, 2)
	case "max":
		pivotTable, err = aggregation.PivotMax(pivotData, 0, 1, 2)
	case "count":
		pivotTable, err = aggregation.PivotCount(pivotData, 0, 1, 2)
	case "median":
		pivotTable, err = aggregation.PivotMedian(pivotData, 0, 1, 2)
	case "stddev":
		pivotTable, err = aggregation.PivotStdDev(pivotData, 0, 1, 2)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown aggregation function"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pivotTable)
}

// Cleaning
func (h *DatasetHandler) DropRowsWithMissingHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Query("datasetID")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset"})
		return
	}

	cleanedRows := cleaning.DropRowsWithMissing(rows)
	c.JSON(http.StatusOK, gin.H{"rows": cleanedRows})
}

func (h *DatasetHandler) FillMissingWithHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Query("datasetID")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	defaultValue := c.Query("defaultValue")
	if defaultValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "defaultValue query param required"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset"})
		return
	}

	cleanedRows := cleaning.FillMissingWith(rows, defaultValue)
	c.JSON(http.StatusOK, gin.H{"rows": cleanedRows})
}

func (h *DatasetHandler) ApplyLogTransformationHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Query("datasetID")
	colStr := c.Query("col")

	if datasetIDStr == "" || colStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datasetID and col are required"})
		return
	}

	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	col, err := strconv.Atoi(colStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column index"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset"})
		return
	}

	transformedRows, err := cleaning.ApplyLogTransformation(rows, col)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rows": transformedRows})
}

func (h *DatasetHandler) NormalizeColumnHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Query("datasetID")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	var req struct {
		Column      int    `json:"column"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dataset rows"})
		return
	}

	normalized, err := cleaning.NormalizeColumn(rows, req.Column)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build Chart.js-compatible response
	var resultRows [][]any
	for i, val := range normalized {
		label := fmt.Sprintf("%v", rows[i][0]) // use column 0 as label (or change as needed)
		resultRows = append(resultRows, []any{label, val})
	}

	_, err = h.Service.UpdateDataset(c, datasetID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dataset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rows": resultRows,
	})
}

func (h *DatasetHandler) StandardizeColumnHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Query("datasetID")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	var req struct {
		Column int `json:"column"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dataset rows"})
		return
	}

	standardized, err := cleaning.StandardizeColumn(rows, req.Column)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Column standardized successfully",
		"data":    standardized,
	})
}

func (h *DatasetHandler) DropColumnsHandler(c *gin.Context) {
	datasetIDStr := c.Param("datasetID")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	var req struct {
		Columns []uuid.UUID `json:"columns"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	for _, fieldID := range req.Columns {
		err := h.Service.DeleteFieldFromDataset(c, datasetID, fieldID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete column with ID %s", fieldID)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Columns dropped successfully"})
}

func (h *DatasetHandler) RenameColumnsHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Param("id")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	var req struct {
		NewHeaders []string `json:"new_headers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dataset rows"})
		return
	}

	cleaned, err := cleaning.RenameColumns(rows, req.NewHeaders)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Columns renamed successfully",
		"data":    cleaned,
	})
}

// Correlation
func TransposeFloat(data [][]float64) [][]float64 {
	if len(data) == 0 {
		return [][]float64{}
	}
	numCols := len(data[0])
	transposed := make([][]float64, numCols)
	for i := range transposed {
		transposed[i] = make([]float64, len(data))
		for j := range data {
			transposed[i][j] = data[j][i]
		}
	}
	return transposed
}

func (h *DatasetHandler) PearsonHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Param("id")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	var req struct {
		XColumn int `json:"x_column"`
		YColumn int `json:"y_column"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var xVals, yVals []float64
	for _, row := range rows {
		if req.XColumn >= len(row) || req.YColumn >= len(row) {
			continue
		}
		x, xErr := strconv.ParseFloat(row[req.XColumn], 64)
		y, yErr := strconv.ParseFloat(row[req.YColumn], 64)
		if xErr == nil && yErr == nil {
			xVals = append(xVals, x)
			yVals = append(yVals, y)
		}
	}

	r, err := correlation.PearsonCorrelation(xVals, yVals)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pearson": r})
}

func (h *DatasetHandler) SpearmanHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Param("id")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	var req struct {
		XColumn int `json:"x_column"`
		YColumn int `json:"y_column"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var xVals, yVals []float64
	for _, row := range rows {
		if req.XColumn >= len(row) || req.YColumn >= len(row) {
			continue
		}
		x, xErr := strconv.ParseFloat(row[req.XColumn], 64)
		y, yErr := strconv.ParseFloat(row[req.YColumn], 64)
		if xErr == nil && yErr == nil {
			xVals = append(xVals, x)
			yVals = append(yVals, y)
		}
	}

	r, err := correlation.SpearmanCorrelation(xVals, yVals)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"spearman": r})
}

func (h *DatasetHandler) CorrelationMatrixHandler(c *gin.Context) {
	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	datasetIDStr := c.Param("id")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	var req struct {
		Columns []int  `json:"columns"`
		Method  string `json:"method"` // "pearson" or "spearman"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Clean up and convert to float64
	clean := make([][]float64, 0, len(rows))
	for _, row := range rows {
		var numericRow []float64
		for _, colIdx := range req.Columns {
			if colIdx >= len(row) {
				numericRow = append(numericRow, 0)
				continue
			}
			val, err := strconv.ParseFloat(row[colIdx], 64)
			if err != nil {
				val = 0
			}
			numericRow = append(numericRow, val)
		}
		clean = append(clean, numericRow)
	}

	transposed := TransposeFloat(clean)

	matrix, err := correlation.CorrelationMatrix(transposed, req.Columns, req.Method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	headers, err := h.Service.GetDatasetHeaders(c, datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get headers"})
		return
	}

	result := make(map[string]map[string]float64)
	for i, row := range matrix {
		headerI := headers[req.Columns[i]]
		if headerI == "" {
			headerI = fmt.Sprintf("Col%d", req.Columns[i])
		}
		result[headerI] = make(map[string]float64)
		for j, val := range row {
			headerJ := headers[req.Columns[j]]
			if headerJ == "" {
				headerJ = fmt.Sprintf("Col%d", req.Columns[j])
			}
			result[headerI][headerJ] = val
		}
	}

	c.JSON(http.StatusOK, result)
}

// Descriptive Stats
func getDatasetAndUserIDFromContext(c *gin.Context) (uuid.UUID, uuid.UUID, error) {
	datasetIDStr := c.Query("dataset_id")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid dataset_id")
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, uuid.Nil, fmt.Errorf("user not authenticated")
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid user ID type")
	}

	return datasetID, userID, nil
}

func (h *AnalyticsHandler) MeanHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	mean, err := descriptives.Mean(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"mean": mean})
}

func (h *AnalyticsHandler) MedianHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	median, err := descriptives.Median(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"median": median})
}

func (h *AnalyticsHandler) ModeHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	headers, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Optional: Validate column exists
	colIndex := -1
	for i, h := range headers {
		if h == column {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Column %q not found", column)})
		return
	}

	var values []string
	for _, row := range rows {
		if colIndex < len(row) && row[colIndex] != "" {
			values = append(values, row[colIndex])
		}
	}

	modes, err := descriptives.Mode(values)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"mode": modes})
}

func (h *AnalyticsHandler) StdDevHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	stddev, err := descriptives.StdDev(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"stddev": stddev})
}

func (h *AnalyticsHandler) VarianceHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	variance, err := descriptives.Variance(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"variance": math.Round(variance*10000) / 10000})
}

func (h *AnalyticsHandler) MinHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	min, err := descriptives.Min(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"min": min})
}

func (h *AnalyticsHandler) MaxHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	max, err := descriptives.Max(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"max": max})
}

func (h *AnalyticsHandler) RangeHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	rng, err := descriptives.Range(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"range": rng})
}

func (h *AnalyticsHandler) SumHandler(c *gin.Context) {
	column := c.Query("column")
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c, datasetID, userID, column)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sum, err := descriptives.Sum(data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"sum": sum})
}

func (h *AnalyticsHandler) CountHandler(c *gin.Context) {
	datasetID, userID, err := getDatasetAndUserIDFromContext(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c, datasetID, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"count": len(rows)})
}

// Distribution
func (h *AnalyticsHandler) HistogramHandler(c *gin.Context) {
	datasetID := c.Query("dataset_id")
	column := c.Query("column")
	numBinsStr := c.DefaultQuery("num_bins", "10")

	if datasetID == "" || column == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing dataset_id or column"})
		return
	}

	id, err := uuid.Parse(datasetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset_id"})
		return
	}

	numBins, err := strconv.Atoi(numBinsStr)
	if err != nil || numBins <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid num_bins value"})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c.Request.Context(), id, userID, column)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to extract column: %v", err)})
		return
	}

	binEdges, binCounts, err := distribution.Histogram(data, numBins)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	labels, counts := distribution.FormatHistogramForChartJS(binEdges, binCounts)

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"counts": counts,
	})
}

func getDatasetIDAndColumnFromQuery(c *gin.Context) (uuid.UUID, string, error) {
	datasetIDStr := c.Query("dataset_id")
	column := c.Query("column")
	if datasetIDStr == "" || column == "" {
		return uuid.Nil, "", errors.New("missing dataset_id or column in query")
	}
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		return uuid.Nil, "", errors.New("invalid dataset_id")
	}
	return datasetID, column, nil
}

func (h *AnalyticsHandler) KDEHandler(c *gin.Context) {
	datasetID, columnName, err := getDatasetIDAndColumnFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	numPoints, _ := strconv.Atoi(c.DefaultQuery("num_points", "100"))
	bandwidth, _ := strconv.ParseFloat(c.DefaultQuery("bandwidth", "1.0"), 64)

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	_, _, err = h.DatasetService.GetDatasetRows(c.Request.Context(), userID, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	numericData, err := h.DatasetService.GetNumericColumnValues(c.Request.Context(), userID, datasetID, columnName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	xs, ys, err := distribution.KDEApproximate(numericData, numPoints, bandwidth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	labels, densities := distribution.FormatKDEForChartJS(xs, ys)

	c.JSON(http.StatusOK, gin.H{
		"labels":    labels,
		"densities": densities,
	})
}

// FilterSort
func (h *DatasetHandler) FilterSortHandler(c *gin.Context) {
	datasetID, _, err := getDatasetIDAndColumnFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	headers, rows, err := h.Service.GetDatasetRows(c.Request.Context(), userID, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filterColumn := c.DefaultQuery("filter_col", "")
	filterOp := c.DefaultQuery("filter_op", "")
	filterValStr := c.DefaultQuery("filter_val", "")
	sortBy := c.DefaultQuery("sort_by", "")
	order := c.DefaultQuery("order", "asc")

	filterVal, _ := strconv.Atoi(filterValStr)

	colIndex := -1
	for i, h := range headers {
		if h == filterColumn {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter column"})
		return
	}

	var filtered [][]string
	for _, row := range rows {
		val, err := strconv.Atoi(row[colIndex])
		if err != nil {
			continue
		}
		if filterOp == "gt" && val > filterVal {
			filtered = append(filtered, row)
		}
	}

	if sortBy != "" {
		sortIndex := -1
		for i, h := range headers {
			if h == sortBy {
				sortIndex = i
				break
			}
		}
		if sortIndex != -1 {
			sort.Slice(filtered, func(i, j int) bool {
				valI, _ := strconv.Atoi(filtered[i][sortIndex])
				valJ, _ := strconv.Atoi(filtered[j][sortIndex])
				if order == "asc" {
					return valI < valJ
				}
				return valI > valJ
			})
		}
	}

	var response []map[string]interface{}
	for _, row := range filtered {
		obj := map[string]interface{}{}
		for i, val := range row {
			obj[headers[i]] = val
		}
		response = append(response, obj)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// Outliers
func (h *AnalyticsHandler) ZScoreOutliersHandler(c *gin.Context) {
	datasetID, columnName, err := getDatasetIDAndColumnFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	threshold, _ := strconv.ParseFloat(c.DefaultQuery("threshold", "2.0"), 64)

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	_, _, err = h.DatasetService.GetDatasetRows(c.Request.Context(), userID, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c.Request.Context(), userID, datasetID, columnName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	indices, err := outliers.ZScoreOutliers(data, threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"indices": indices})
}

func (h *AnalyticsHandler) IQROutliersHandler(c *gin.Context) {
	datasetID, columnName, err := getDatasetIDAndColumnFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	_, _, err = h.DatasetService.GetDatasetRows(c.Request.Context(), userID, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c.Request.Context(), userID, datasetID, columnName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	indices, err := outliers.IQROutliers(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"indices": indices})
}

func (h *AnalyticsHandler) BoxPlotHandler(c *gin.Context) {
	datasetID, columnName, err := getDatasetIDAndColumnFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	_, _, err = h.DatasetService.GetDatasetRows(c.Request.Context(), userID, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := h.DatasetService.GetNumericColumnValues(c.Request.Context(), userID, datasetID, columnName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Q1, Q3, IQR, lower, upper, err := outliers.BoxPlotData(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	labels, values := outliers.FormatBoxPlotForChartJS(Q1, Q3, lower, upper)

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"values": values,
		"stats": gin.H{
			"Q1":            Q1,
			"Q3":            Q3,
			"IQR":           IQR,
			"lower_outlier": lower,
			"upper_outlier": upper,
		},
	})
}

func Transpose(data [][]string) [][]string {
	if len(data) == 0 {
		return [][]string{}
	}
	numRows := len(data)
	numCols := len(data[0])
	result := make([][]string, numCols)
	for i := 0; i < numCols; i++ {
		result[i] = make([]string, numRows)
		for j := 0; j < numRows; j++ {
			result[i][j] = data[j][i]
		}
	}
	return result
}
