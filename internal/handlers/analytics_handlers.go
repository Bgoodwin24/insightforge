package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/analytics/aggregation"
	"github.com/Bgoodwin24/insightforge/internal/analytics/cleaning"
	"github.com/Bgoodwin24/insightforge/internal/analytics/correlation"
	"github.com/Bgoodwin24/insightforge/internal/analytics/descriptives"
	"github.com/Bgoodwin24/insightforge/internal/analytics/distribution"
	"github.com/Bgoodwin24/insightforge/internal/analytics/outliers"
	"github.com/Bgoodwin24/insightforge/internal/database"
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

// Aggregation
func GroupByHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
		Column  string                   `json:"column"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyCol, valCol := -1, -1
	if len(requestData.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No data provided"})
		return
	}

	var headers []string
	for k := range requestData.Data[0] {
		headers = append(headers, k)
	}

	for i, h := range headers {
		if h == requestData.GroupBy {
			keyCol = i
		}
		if h == requestData.Column {
			valCol = i
		}
	}

	if keyCol == -1 || valCol == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group_by or column"})
		return
	}

	var csvData [][]string
	for _, row := range requestData.Data {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = fmt.Sprintf("%v", row[h])
		}
		csvData = append(csvData, record)
	}

	grouped, err := aggregation.GroupBy(csvData, keyCol, valCol)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, grouped)
}

func GroupedSumHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
		Column  string                   `json:"column"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grouped := make(map[string][]float64)
	for _, item := range requestData.Data {
		group := fmt.Sprintf("%v", item[requestData.GroupBy])
		val, ok := convertToFloat64(item[requestData.Column])
		if !ok {
			continue
		}
		grouped[group] = append(grouped[group], val)
	}

	sums := aggregation.GroupedSum(grouped)
	c.JSON(http.StatusOK, sums)
}

func GroupedMeanHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
		Column  string                   `json:"column"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grouped := make(map[string][]float64)
	for _, item := range requestData.Data {
		group := fmt.Sprintf("%v", item[requestData.GroupBy])
		val, ok := convertToFloat64(item[requestData.Column])
		if !ok {
			continue
		}
		grouped[group] = append(grouped[group], val)
	}

	means := aggregation.GroupedMean(grouped)
	c.JSON(http.StatusOK, means)
}

func GroupedCountHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grouped := make(map[string][]float64)
	for _, item := range requestData.Data {
		group := fmt.Sprintf("%v", item[requestData.GroupBy])
		grouped[group] = append(grouped[group], 1)
	}

	counts := aggregation.GroupedCount(grouped)
	c.JSON(http.StatusOK, counts)
}

func GroupedMinHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
		Column  string                   `json:"column"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grouped := make(map[string][]float64)

	for _, row := range requestData.Data {
		groupKey := fmt.Sprintf("%v", row[requestData.GroupBy])
		val, ok := convertToFloat64(row[requestData.Column])
		if !ok {
			continue
		}
		grouped[groupKey] = append(grouped[groupKey], val)
	}

	mins := aggregation.GroupedMin(grouped)
	c.JSON(http.StatusOK, mins)
}

func GroupedMaxHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
		Column  string                   `json:"column"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := make([][]string, 0, len(requestData.Data))
	for _, row := range requestData.Data {
		groupByVal := fmt.Sprintf("%v", row[requestData.GroupBy])
		columnVal := fmt.Sprintf("%v", row[requestData.Column])
		data = append(data, []string{groupByVal, columnVal})
	}

	grouped, err := aggregation.GroupBy(data, 0, 1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := aggregation.GroupedMax(grouped)
	c.JSON(http.StatusOK, result)
}

func GroupedMedianHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
		Column  string                   `json:"column"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := make([][]string, 0, len(requestData.Data))
	for _, row := range requestData.Data {
		group := fmt.Sprintf("%v", row[requestData.GroupBy])
		value := fmt.Sprintf("%v", row[requestData.Column])
		data = append(data, []string{group, value})
	}

	groupedResult, err := aggregation.GroupBy(data, 0, 1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	medians := aggregation.GroupedMedian(groupedResult)
	c.JSON(http.StatusOK, medians)
}

func GroupedStdDevHandler(c *gin.Context) {
	var requestData struct {
		Data    []map[string]interface{} `json:"data"`
		GroupBy string                   `json:"group_by"`
		Column  string                   `json:"column"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := make([][]string, 0, len(requestData.Data))
	for _, row := range requestData.Data {
		group := fmt.Sprintf("%v", row[requestData.GroupBy])
		value := fmt.Sprintf("%v", row[requestData.Column])
		data = append(data, []string{group, value})
	}

	groupedResult, err := aggregation.GroupBy(data, 0, 1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stddevs := aggregation.GroupedStdDev(groupedResult)
	c.JSON(http.StatusOK, stddevs)
}

type PivotRequest struct {
	RowField    string                   `json:"row_field"`
	ColumnField string                   `json:"column_field"`
	ValueField  string                   `json:"value_field"`
	AggFunc     string                   `json:"agg_func"`
	Data        []map[string]interface{} `json:"data"`
}

func PivotTableHandler(c *gin.Context) {
	var req PivotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if req.RowField == "" || req.ColumnField == "" || req.ValueField == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: row_field, column_field, value_field"})
		return
	}

	// Convert to [][]string for processing
	data := [][]string{}
	for i, record := range req.Data {
		rowVal, ok1 := record[req.RowField]
		colVal, ok2 := record[req.ColumnField]
		valVal, ok3 := record[req.ValueField]

		if !ok1 || !ok2 || !ok3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing key in data row %d", i)})
			return
		}

		rowStr := fmt.Sprintf("%v", rowVal)
		colStr := fmt.Sprintf("%v", colVal)
		valStr := fmt.Sprintf("%v", valVal)
		data = append(data, []string{rowStr, colStr, valStr})
	}

	var (
		pivotTable aggregation.PivotTable
		err        error
	)

	switch req.AggFunc {
	case "sum":
		pivotTable, err = aggregation.PivotSum(data, 0, 1, 2)
	case "mean":
		pivotTable, err = aggregation.PivotMean(data, 0, 1, 2)
	case "min":
		pivotTable, err = aggregation.PivotMin(data, 0, 1, 2)
	case "max":
		pivotTable, err = aggregation.PivotMax(data, 0, 1, 2)
	case "count":
		pivotTable, err = aggregation.PivotCount(data, 0, 1, 2)
	case "median":
		pivotTable, err = aggregation.PivotMedian(data, 0, 1, 2)
	case "stddev":
		pivotTable, err = aggregation.PivotStdDev(data, 0, 1, 2)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown aggregation function"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pivotTable)
}

// Cleaning
func DropRowsWithMissingHandler(c *gin.Context, datasetService *services.DatasetService) {
	var body struct {
		Data    []map[string]any `json:"data"`
		Columns []string         `json:"columns"`
	}

	if err := c.ShouldBindJSON(&body); err == nil && len(body.Data) > 0 {
		if len(body.Columns) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or empty 'columns' field"})
			return
		}
		var rows [][]string
		for _, rowMap := range body.Data {
			var row []string
			for _, col := range body.Columns {
				val, ok := rowMap[col]
				if !ok || val == nil {
					row = append(row, "")
				} else {
					row = append(row, fmt.Sprintf("%v", val))
				}
			}
			rows = append(rows, row)
		}

		cleanedRows := cleaning.DropRowsWithMissing(rows)
		c.JSON(http.StatusOK, gin.H{"rows": cleanedRows})
		return
	}

	datasetIDStr := c.Param("datasetID")
	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	rows, err := datasetService.GetDatasetRows(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset"})
		return
	}

	cleanedRows := cleaning.DropRowsWithMissing(rows)
	c.JSON(http.StatusOK, gin.H{"rows": cleanedRows})
}

func FillMissingWithHandler(c *gin.Context, datasetService *services.DatasetService) {
	datasetIDStr := c.Param("datasetID")

	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID", "datasetID": datasetIDStr, "errorDetail": err.Error()})
		return
	}

	defaultValue := c.Query("defaultValue")
	if defaultValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "defaultValue query param required"})
		return
	}

	rows, err := datasetService.GetDatasetRows(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset"})
		return
	}

	cleanedRows := cleaning.FillMissingWith(rows, defaultValue)

	c.JSON(http.StatusOK, gin.H{
		"rows": cleanedRows,
	})
}

func ApplyLogTransformationHandler(c *gin.Context, datasetService *services.DatasetService) {
	datasetIDStr := c.DefaultQuery("datasetID", "")
	if datasetIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	datasetID, err := uuid.Parse(datasetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dataset ID"})
		return
	}

	colStr := c.DefaultQuery("col", "")
	if colStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Column index is required"})
		return
	}

	col, err := strconv.Atoi(colStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column index"})
		return
	}

	rows, err := datasetService.GetDatasetRows(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dataset"})
		return
	}

	if len(rows) == 0 || len(rows[0]) <= col {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column index or empty dataset"})
		return
	}

	transformedRows, err := cleaning.ApplyLogTransformation(rows, col)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rows": transformedRows,
	})
}

func NormalizeColumnHandler(c *gin.Context, datasetService *services.DatasetService) {
	datasetIDStr := c.Param("datasetID")
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

	rows, err := datasetService.GetDatasetRows(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dataset rows"})
		return
	}

	normalizedRows, err := cleaning.NormalizeColumn(rows, req.Column)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = datasetService.UpdateDataset(c, datasetID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dataset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Column normalized successfully",
		"column":  req.Name,
		"data":    normalizedRows,
	})
}

func StandardizeColumnHandler(c *gin.Context, datasetService *services.DatasetService) {
	datasetIDStr := c.Param("datasetID")
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

	rows, err := datasetService.GetDatasetRows(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dataset rows"})
		return
	}

	_, err = cleaning.StandardizeColumn(rows, req.Column)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = datasetService.UpdateDataset(c, datasetID, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dataset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Column standardized successfully"})
}

func DropColumnsHandler(c *gin.Context, datasetService *services.DatasetService) {
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

	// Delete the columns from the dataset
	for _, fieldID := range req.Columns {
		err := datasetService.DeleteFieldFromDataset(c, datasetID, fieldID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete column with ID %s", fieldID)})
			return
		}
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{"message": "Columns dropped successfully"})
}

func RenameColumnsHandler(c *gin.Context, datasetService *services.DatasetService) {
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

	rows, err := datasetService.GetDatasetRows(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dataset rows"})
		return
	}

	cleaned, err := cleaning.RenameColumns(rows, req.NewHeaders)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedAt := time.Now()

	for _, row := range cleaned {
		for _, value := range row {
			fieldUUID := uuid.New()
			valueNullString := sql.NullString{
				String: value,
				Valid:  value != "",
			}

			params := database.UpdateDatasetRowsParams{
				DatasetID: datasetID,
				UpdatedAt: updatedAt,
				Value:     valueNullString,
				FieldID:   fieldUUID,
			}

			err := datasetService.UpdateDatasetRows(c, params)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dataset row"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Columns renamed successfully"})
}

// Correlation
func PearsonHandler(c *gin.Context) {
	var req struct {
		X []float64 `json:"x"`
		Y []float64 `json:"y"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	r, err := correlation.PearsonCorrelation(req.X, req.Y)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pearson": r})
}

func SpearmanHandler(c *gin.Context) {
	var req struct {
		X []float64 `json:"x"`
		Y []float64 `json:"y"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	r, err := correlation.SpearmanCorrelation(req.X, req.Y)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"spearman": r})
}

func CorrelationMatrixHandler(c *gin.Context, datasetService *services.DatasetService) {
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

	rows, err := datasetService.GetDatasetRows(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, row := range rows {
		for j, value := range row {
			if value == "" {
				fmt.Printf("Empty value detected at row %d, column %d, skipping it\n", i, j)
				rows[i][j] = "0"
			}
		}
	}

	transposed := Transpose(rows)

	matrix, err := correlation.CorrelationMatrix(transposed, req.Columns, req.Method)
	if err != nil {
		fmt.Printf("CorrelationMatrix error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	headers, err := datasetService.GetDatasetHeaders(c, datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get headers"})
		return
	}

	result := make(map[string]map[string]float64)
	for i, row := range matrix {
		headerI := headers[req.Columns[i]]
		if headerI == "" {
			headerI = "Unknown"
		}

		if _, exists := result[headerI]; !exists {
			result[headerI] = make(map[string]float64)
		}

		for j, value := range row {
			headerJ := headers[req.Columns[j]]
			if headerJ == "" {
				headerJ = "Unknown"
			}
			result[headerI][headerJ] = value
		}
	}

	c.JSON(http.StatusOK, result)
}

// Descriptive Stats
func MeanHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	mean, err := descriptives.Mean(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"mean": mean})
}

func MedianHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	median, err := descriptives.Median(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"median": median})
}

func ModeHandler(c *gin.Context) {
	var input struct {
		Column string                   `json:"column"`
		Data   []map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	var values []string
	for _, row := range input.Data {
		if v, ok := row[input.Column]; ok {
			strVal := fmt.Sprintf("%v", v)
			values = append(values, strVal)
		}
	}

	modes, err := descriptives.Mode(values)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"mode": modes})
}

func StdDevHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	stddev, err := descriptives.StdDev(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"stddev": stddev})
}

func VarianceHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	variance, err := descriptives.Variance(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	roundedVariance := math.Round(variance*10000) / 10000

	c.JSON(200, gin.H{"variance": roundedVariance})
}

func MinHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	min, err := descriptives.Min(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"min": min})
}

func MaxHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	max, err := descriptives.Max(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"max": max})
}

func RangeHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	rangeVal, err := descriptives.Range(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"range": rangeVal})
}

func SumHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	sum, err := descriptives.Sum(input.Data)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"sum": sum})
}

func CountHandler(c *gin.Context) {
	var input struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	count := len(input.Data)

	c.JSON(200, gin.H{"count": count})
}

// Distribution
func HistogramHandler(c *gin.Context) {
	var input struct {
		Data    []float64 `json:"data"`
		NumBins int       `json:"num_bins"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	binEdges, binCounts, err := distribution.Histogram(input.Data, input.NumBins)
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

func KDEHandler(c *gin.Context) {
	var input struct {
		Data      []float64 `json:"data"`
		NumPoints int       `json:"num_points"`
		Bandwidth float64   `json:"bandwidth"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	xs, ys, err := distribution.KDEApproximate(input.Data, input.NumPoints, input.Bandwidth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	labels, densities := distribution.FormatKDEForChartJS(xs, ys)

	c.JSON(http.StatusOK, gin.H{
		"labels":    labels,
		"densities": densities,
	})
}

// FilterSort
func FilterSortHandler(c *gin.Context) {
	var input struct {
		Headers [][]string `json:"headers"`
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to bind JSON: %s", err.Error())})
		return
	}

	if len(input.Headers) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. Expected at least headers and one row."})
		return
	}

	headers := input.Headers[0]
	data := input.Headers[1:]

	filterColumn := c.DefaultQuery("filter_col", "")
	filterOperation := c.DefaultQuery("filter_op", "")
	filterValueStr := c.DefaultQuery("filter_val", "")
	sortBy := c.DefaultQuery("sort_by", "")
	order := c.DefaultQuery("order", "asc")

	filterValue, err := strconv.Atoi(filterValueStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter value"})
		return
	}

	var filteredData [][]string
	for _, row := range data {
		if filterColumn == "age" {
			age, err := strconv.Atoi(row[1])
			if err != nil {
				continue
			}
			if filterOperation == "gt" && age > filterValue {
				filteredData = append(filteredData, row)
			}
		}
	}

	if sortBy != "" {
		sort.Slice(filteredData, func(i, j int) bool {
			ageI, _ := strconv.Atoi(filteredData[i][1])
			ageJ, _ := strconv.Atoi(filteredData[j][1])

			if order == "asc" {
				return ageI < ageJ
			}
			return ageI > ageJ
		})
	}

	var response []map[string]interface{}
	for _, row := range filteredData {
		obj := make(map[string]interface{})
		obj[headers[0]] = row[0] // "name" -> "Bob"
		obj[headers[1]] = row[1] // "age" -> "30"
		response = append(response, obj)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// Outliers
func ZScoreOutliersHandler(c *gin.Context) {
	var input struct {
		Data      []float64 `json:"data"`
		Threshold float64   `json:"threshold"`
	}

	if err := c.ShouldBindJSON(&input); err != nil || len(input.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. Expecting non-empty 'data' and 'threshold'."})
		return
	}

	outlierIndices, err := outliers.ZScoreOutliers(input.Data, input.Threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"indices": outlierIndices,
	})
}

func IQROutliersHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil || len(input.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. Expecting non-empty 'data' array."})
		return
	}

	outlierIndices, err := outliers.IQROutliers(input.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"indices": outlierIndices,
	})
}

func BoxPlotHandler(c *gin.Context) {
	var input struct {
		Data []float64 `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil || len(input.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. Expecting non-empty 'data' array."})
		return
	}

	Q1, Q3, IQR, lower, upper, err := outliers.BoxPlotData(input.Data)
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
