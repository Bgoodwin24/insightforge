package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/handlers"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func performHandlerTest(t *testing.T, handlerFunc gin.HandlerFunc, url string, jsonBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handlerFunc(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", w.Code)
	}

	return w
}

func performServiceHandlerTest(t *testing.T, handler func(*gin.Context, *services.DatasetService), datasetService *services.DatasetService, url string, jsonBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	urlWithoutQuery := strings.Split(url, "?")[0]
	parts := strings.Split(urlWithoutQuery, "/")
	if len(parts) > 0 {
		c.Params = append(c.Params, gin.Param{Key: "datasetID", Value: parts[len(parts)-1]})
	}

	handler(c, datasetService)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", w.Code)
	}

	return w
}

func newMockDatasetService() *services.DatasetService {
	return &services.DatasetService{}
}

func TestGroupByHandler(t *testing.T) {
	body := `{
		"data": [
			{"category": "A", "value": 10},
			{"category": "A", "value": 15},
			{"category": "B", "value": 5}
		],
		"group_by": "category",
		"column": "value"
	}`

	req := httptest.NewRequest(http.MethodPost, "/analytics/groupby", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handlers.GroupByHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.JSONEq(t, `{"A": [10, 15], "B": [5]}`, w.Body.String())
}

func TestGroupedSumHandler(t *testing.T) {
	body := `{
		"group_by": "category",
		"column": "amount",
		"data": [
			{"category": "A", "amount": 10},
			{"category": "A", "amount": 15},
			{"category": "B", "amount": 5}
		]
	}`

	w := performHandlerTest(t, handlers.GroupedSumHandler, "/analytics/grouped/sum", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"A":25`)
	assert.Contains(t, w.Body.String(), `"B":5`)
}

func TestGroupedMeanHandler(t *testing.T) {
	body := `{
		"group_by": "type",
		"column": "value",
		"data": [
			{"type": "X", "value": 10},
			{"type": "X", "value": 30},
			{"type": "Y", "value": 20}
		]
	}`

	w := performHandlerTest(t, handlers.GroupedMeanHandler, "/analytics/grouped/mean", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"X":20`)
	assert.Contains(t, w.Body.String(), `"Y":20`)
}

func TestGroupedCountHandler(t *testing.T) {
	body := `{
		"group_by": "group",
		"data": [
			{"group": "G1"},
			{"group": "G1"},
			{"group": "G2"}
		]
	}`

	w := performHandlerTest(t, handlers.GroupedCountHandler, "/analytics/grouped/count", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"G1":2`)
	assert.Contains(t, w.Body.String(), `"G2":1`)
}

func TestGroupedMinHandler(t *testing.T) {
	body := `{
		"group_by": "category",
		"column": "score",
		"data": [
			{"category": "alpha", "score": 9},
			{"category": "alpha", "score": 3},
			{"category": "beta", "score": 7}
		]
	}`

	w := performHandlerTest(t, handlers.GroupedMinHandler, "/analytics/grouped/min", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"alpha":3`)
	assert.Contains(t, w.Body.String(), `"beta":7`)
}

func TestGroupedMaxHandler(t *testing.T) {
	body := `{
		"group_by": "dept",
		"column": "salary",
		"data": [
			{"dept": "HR", "salary": 50000},
			{"dept": "HR", "salary": 70000},
			{"dept": "IT", "salary": 80000}
		]
	}`

	w := performHandlerTest(t, handlers.GroupedMaxHandler, "/analytics/grouped/max", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"HR":70000`)
	assert.Contains(t, w.Body.String(), `"IT":80000`)
}

func TestGroupedMedianHandler(t *testing.T) {
	body := `{
		"group_by": "class",
		"column": "score",
		"data": [
			{"class": "A", "score": 10},
			{"class": "A", "score": 30},
			{"class": "A", "score": 20},
			{"class": "B", "score": 5},
			{"class": "B", "score": 15}
		]
	}`

	w := performHandlerTest(t, handlers.GroupedMedianHandler, "/analytics/grouped/median", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"A":20`)
	assert.Contains(t, w.Body.String(), `"B":10`)
}

func TestGroupedStdDevHandler(t *testing.T) {
	body := `{
		"group_by": "team",
		"column": "score",
		"data": [
			{"team": "Red", "score": 10},
			{"team": "Red", "score": 20},
			{"team": "Red", "score": 30},
			{"team": "Blue", "score": 15},
			{"team": "Blue", "score": 15}
		]
	}`

	w := performHandlerTest(t, handlers.GroupedStdDevHandler, "/analytics/grouped/stddev", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"Red"`)
	assert.Contains(t, w.Body.String(), `"Blue":0`)
}

func TestPivotTableHandler(t *testing.T) {
	body := `{
		"row_field": "region",
		"column_field": "product",
		"value_field": "sales",
		"agg_func": "sum",
		"data": [
			{"region": "North", "product": "A", "sales": 100},
			{"region": "North", "product": "B", "sales": 150},
			{"region": "South", "product": "A", "sales": 200}
		]
	}`

	w := performHandlerTest(t, handlers.PivotTableHandler, "/analytics/pivot", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"North"`)
	assert.Contains(t, w.Body.String(), `"A":100`)
	assert.Contains(t, w.Body.String(), `"B":150`)
	assert.Contains(t, w.Body.String(), `"South"`)
	assert.Contains(t, w.Body.String(), `"A":200`)
}

func TestDropRowsWithMissingHandler(t *testing.T) {
	service := newMockDatasetService()
	body := `{
		"columns": ["score", "age"],
		"data": [
			{"score": 85, "age": 30},
			{"score": null, "age": 22},
			{"score": 90, "age": null},
			{"score": 95, "age": 28}
		]
	}`

	w := performServiceHandlerTest(t, handlers.DropRowsWithMissingHandler, service, "/analytics/missing/drop", body)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `[["85","30"],["95","28"]]`)

	assert.NotContains(t, w.Body.String(), `[["null","22"]]`)
	assert.NotContains(t, w.Body.String(), `[["90","null"]]`)
}

func TestFillMissingWithHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	testutils.CleanDB(repo)
	service := &services.DatasetService{Repo: repo}

	user := testutils.CreateTestUser(t, repo, "test@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Missing Fill Dataset", "Test dataset for filling missing values")

	fieldID := uuid.New()
	now := time.Now()
	_, err := repo.DB.Exec(`
        INSERT INTO dataset_fields (id, dataset_id, name, data_type, created_at)
        VALUES ($1, $2, 'score', 'numeric', $3)
    `, fieldID, dataset.ID, now)
	require.NoError(t, err)

	val10 := "10"
	testutils.InsertTestRecord(t, repo, dataset.ID, fieldID, "")
	testutils.InsertTestRecord(t, repo, dataset.ID, fieldID, val10)
	testutils.InsertTestRecord(t, repo, dataset.ID, fieldID, "")

	body := `{
		"column": "score",
		"value": 0
	}`

	url := "/analytics/cleaning/fill-missing-with/" + dataset.ID.String() + "?defaultValue=0"
	w := performServiceHandlerTest(t, handlers.FillMissingWithHandler, service, url, body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"rows":[["0"],["10"],["0"]]}`, w.Body.String())
}

func TestApplyLogTransformationHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)

	user := testutils.CreateTestUser(t, repo, "user@example.com")

	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Test Dataset", "A test dataset")

	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")

	values := []string{"1", "10", "100"}
	for _, val := range values {
		testutils.InsertTestRecord(t, repo, dataset.ID, fieldID, val)
	}

	url := "/analytics/transform/log?datasetID=" + dataset.ID.String() + "&col=0"

	body := `{
        "column": "value",
        "data": [
            {"value": 1},
            {"value": 10},
            {"value": 100}
        ]
    }`

	w := performServiceHandlerTest(t, handlers.ApplyLogTransformationHandler, service, url, body)

	assert.Equal(t, http.StatusOK, w.Code)

	expected := `{"rows":[["0"],["2.302585092994046"],["4.605170185988092"]]}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestNormalizeColumnHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	email := fmt.Sprintf("testuser+%d@example.com", time.Now().UnixNano())

	user := testutils.CreateTestUser(t, repo, email)
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Test Dataset", "A test dataset")

	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")

	values := []string{"50", "100", "150"}
	for _, val := range values {
		testutils.InsertTestRecord(t, repo, dataset.ID, fieldID, val)
	}

	url := "/analytics/transform/normalize/" + dataset.ID.String()

	body := `{
        "column": 0,
        "name": "Normalized Score",
        "description": "Normalized scores of the dataset"
    }`

	w := performServiceHandlerTest(t, handlers.NormalizeColumnHandler, service, url, body)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"column":"Normalized Score"`)

	assert.Contains(t, w.Body.String(), `"data":[["0.000000"],["0.500000"],["1.000000"]]`)
}

func TestStandardizeColumnHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	testutils.CleanDB(repo)
	service := &services.DatasetService{Repo: repo}
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	user := testutils.CreateTestUser(t, repo, email)
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Test Dataset", "Dataset for standardizing column")

	fieldID := uuid.New()
	now := time.Now()
	_, err := repo.DB.Exec(`
		INSERT INTO dataset_fields (id, dataset_id, name, data_type, created_at)
		VALUES ($1, $2, 'score', 'numeric', $3)
	`, fieldID, dataset.ID, now)
	require.NoError(t, err)

	values := []string{"50", "100", "150"}
	for _, val := range values {
		testutils.InsertTestRecord(t, repo, dataset.ID, fieldID, val)
	}

	body := `{
		"column": 0,
		"name": "Standardized Score",
		"description": "Standardized scores of the dataset"
	}`

	url := "/analytics/cleaning/standardize-column/" + dataset.ID.String()

	w := performServiceHandlerTest(t, handlers.StandardizeColumnHandler, service, url, body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"Column standardized successfully"`)
}

func TestDropColumnsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup and clean the database
	repo := testutils.SetupTestRepo()
	testutils.CleanDB(repo)

	// Create the service and user
	service := services.NewDatasetService(repo)
	email := fmt.Sprintf("testuser+%d@example.com", time.Now().UnixNano())
	user := testutils.CreateTestUser(t, repo, email)

	// Create a dataset for the test
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Test Dataset", "For column drop test")

	// Create two field IDs to test with
	scoreFieldID := uuid.New()
	ageFieldID := uuid.New()

	// Insert fields into the dataset
	testutils.InsertTestField(t, repo, dataset.ID, scoreFieldID, "score", "numeric")
	testutils.InsertTestField(t, repo, dataset.ID, ageFieldID, "age", "numeric")

	// Add one record for both fields
	recordID := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, scoreFieldID, "95")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, ageFieldID, "30")

	// Ensure the fields are inserted by querying them
	params := database.GetDatasetFieldParams{
		ID:        scoreFieldID,
		DatasetID: dataset.ID,
	}

	// Ensure the fields exist before dropping
	_, err := repo.Queries.GetDatasetField(context.Background(), params)
	require.NoError(t, err)

	// Ensure the second field exists
	params.ID = ageFieldID
	_, err = repo.Queries.GetDatasetField(context.Background(), params)
	require.NoError(t, err)

	// Create request body with field IDs to drop
	body := fmt.Sprintf(`{
		"columns": ["%s", "%s"]
	}`, scoreFieldID.String(), ageFieldID.String())

	// Perform the handler test
	w := performServiceHandlerTest(t, handlers.DropColumnsHandler, service, "/analytics/transform/drop-columns/"+dataset.ID.String(), body)

	// Assert the response status is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the fields were successfully dropped by querying them again
	params.ID = scoreFieldID
	_, err = repo.Queries.GetDatasetField(context.Background(), params)
	require.Error(t, err) // Expect error because the field should no longer exist

	// Now check for the second field (ageFieldID)
	params.ID = ageFieldID
	_, err = repo.Queries.GetDatasetField(context.Background(), params)
	require.Error(t, err) // Expect error because the field should no longer exist
}

func TestRenameColumn(t *testing.T) {
	repo := testutils.SetupTestRepo()

	user := testutils.CreateTestUser(t, repo, "testuser@example.com")

	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Test Dataset", "Test dataset description")

	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "old_column", "string")

	newColumnName := "new_column"
	_, err := repo.DB.Exec(`
        UPDATE dataset_fields
        SET name = $1
        WHERE dataset_id = $2 AND name = $3
    `, newColumnName, dataset.ID, "old_column")
	require.NoError(t, err)

	var columnName string
	err = repo.DB.QueryRow(`
        SELECT name
        FROM dataset_fields
        WHERE dataset_id = $1 AND name = $2
    `, dataset.ID, newColumnName).Scan(&columnName)
	require.NoError(t, err)
	require.Equal(t, newColumnName, columnName)
}

func TestPearsonHandler(t *testing.T) {
	body := `{
		"x": [1, 2, 3],
		"y": [2, 4, 6]
	}`

	w := performHandlerTest(t, handlers.PearsonHandler, "/analytics/correlation/pearson", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"pearson":1`)
}

func TestSpearmanHandler(t *testing.T) {
	body := `{
		"x": [1, 2, 3],
		"y": [3, 2, 1]
	}`

	w := performHandlerTest(t, handlers.SpearmanHandler, "/analytics/correlation/spearman", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"spearman":-1`)
}

func TestCorrelationMatrixHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	defer repo.DB.Close()
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "test-corr@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Correlation Dataset", "Test")

	fieldAID := uuid.New()
	fieldBID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldAID, "a", "number")
	testutils.InsertTestField(t, repo, dataset.ID, fieldBID, "b", "number")

	vals := [][]string{
		{"1", "2"},
		{"2", "4"},
	}
	for _, pair := range vals {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldAID, pair[0])
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldBID, pair[1])
	}

	ds := &services.DatasetService{Repo: repo}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	ctx.Params = gin.Params{gin.Param{Key: "id", Value: dataset.ID.String()}}

	ctx.Set("userID", user.ID)
	ctx.Set("email", user.Email)

	body := `{"columns": [0, 1], "method": "pearson"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	handlers.CorrelationMatrixHandler(ctx, ds)

	require.Equal(t, http.StatusOK, w.Code, "Expected status code 200, got %v", w.Code)

	if w.Code == http.StatusInternalServerError {
		fmt.Printf("Error Response: %s\n", w.Body.String())
	}

	var result map[string]map[string]float64
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	require.Contains(t, result, "a")
	require.Contains(t, result["a"], "b")
	require.InDelta(t, 1.0, result["a"]["b"], 0.05)
}

func TestMeanHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := `{
		"data": [10, 20, 30]
	}`

	req, _ := http.NewRequest(http.MethodPost, "/analytics/descriptives/mean", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handlers.MeanHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"mean":20`)
}

func TestMedianHandler(t *testing.T) {
	body := `{
		"data": [10, 20, 30, 40, 50]
	}`

	w := performHandlerTest(t, handlers.MedianHandler, "/analytics/descriptives/median", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"median":30`)
}

func TestModeHandler(t *testing.T) {
	body := `{
		"column": "color",
		"data": [
			{"color": "red"},
			{"color": "blue"},
			{"color": "red"},
			{"color": "green"},
			{"color": "red"}
		]
	}`

	w := performHandlerTest(t, handlers.ModeHandler, "/analytics/descriptives/mode", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"mode":"red"`)
}

func TestStdDevHandler(t *testing.T) {
	body := `{
		"data": [25, 30, 35]
	}`

	w := performHandlerTest(t, handlers.StdDevHandler, "/analytics/descriptives/stddev", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"stddev":5`)
}

func TestVarianceHandler(t *testing.T) {
	body := `{
		"data": [160, 170, 180]
	}`

	w := performHandlerTest(t, handlers.VarianceHandler, "/analytics/descriptives/variance", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"variance":100`)
}

func TestMinHandler(t *testing.T) {
	body := `{
	"data": [90, 85, 70]
	}`

	w := performHandlerTest(t, handlers.MinHandler, "/analytics/descriptives/min", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"min":70`)
}

func TestMaxHandler(t *testing.T) {
	body := `{
	"data": [90, 85, 100]
	}`

	w := performHandlerTest(t, handlers.MaxHandler, "/analytics/descriptives/max", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"max":100`)
}

func TestRangeHandler(t *testing.T) {
	body := `{
	"data": [10, 50, 30]
	}`

	w := performHandlerTest(t, handlers.RangeHandler, "/analytics/descriptives/range", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"range":40`)
}

func TestSumHandler(t *testing.T) {
	body := `{
	"data": [10, 20, 30]
	}`

	w := performHandlerTest(t, handlers.SumHandler, "/analytics/descriptives/sum", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"sum":60`)
}

func TestCountHandler(t *testing.T) {
	body := `{
		"data": [
			{"name": "Alice"},
			{"name": "Bob"},
			{"name": "Alice"}
		]
	}`

	w := performHandlerTest(t, handlers.CountHandler, "/analytics/descriptives/count", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"count":3`)
}

func TestHistogramHandler(t *testing.T) {
	body := `{
		"data": [1, 2, 3, 4, 5, 6], 
		"num_bins": 3
	}`

	w := performHandlerTest(t, handlers.HistogramHandler, "/analytics/distribution/histogram", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"labels"`)
	assert.Contains(t, w.Body.String(), `"counts"`)
}

func TestKDEHandler(t *testing.T) {
	body := `{
		"data": [1, 2, 3, 4, 5],
		"num_points": 10,
		"bandwidth": 1.0
	}`

	w := performHandlerTest(t, handlers.KDEHandler, "/analytics/distribution/kde", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"labels"`)
	assert.Contains(t, w.Body.String(), `"densities"`)
}

func TestFilterSortHandler(t *testing.T) {
	body := `{
		"headers": [
			["name", "age"],
			["Alice", "24"],
			["Bob", "30"],
			["Charlie", "35"]
		]
	}`

	query := "?filter_col=age&filter_op=gt&filter_val=25&sort_by=age&order=asc"

	w := performHandlerTest(t, handlers.FilterSortHandler, "/analytics/filter-sort"+query, body)

	w.Header().Set("Content-Type", "application/json")

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `{"age":"30","name":"Bob"}`)
	assert.Contains(t, w.Body.String(), `{"age":"35","name":"Charlie"}`)
}

func TestZScoreOutliersHandler(t *testing.T) {
	body := `{
		"data": [10, 15, 100, 20],
		"threshold": 1.4
	}`

	w := performHandlerTest(t, handlers.ZScoreOutliersHandler, "/analytics/outliers/zscore", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"indices":[2]`)
}

func TestIQROutliersHandler(t *testing.T) {
	body := `{
		"data": [10, 15, 25, 100, 30]
	}`

	w := performHandlerTest(t, handlers.IQROutliersHandler, "/analytics/outliers/iqr", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"indices":`)
}

func TestBoxPlotHandler(t *testing.T) {
	body := `{
		"data": [10, 15, 25, 100, 30]
	}`

	w := performHandlerTest(t, handlers.BoxPlotHandler, "/analytics/distribution/boxplot", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"stats"`)
}
