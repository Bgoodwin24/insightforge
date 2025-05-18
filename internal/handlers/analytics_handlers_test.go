package handlers_test

import (
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

func TestGroupDatasetBy(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	// Create test user and dataset
	user := testutils.CreateTestUser(t, repo, "groupbyuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Sales", "GroupBy Test")

	// Insert test fields
	categoryFieldID := uuid.New()
	valueFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, categoryFieldID, "category", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	// Insert test records
	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, categoryFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, valueFieldID, "10")

	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, categoryFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, valueFieldID, "15")

	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, categoryFieldID, "B")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, valueFieldID, "5")

	// Create test request context manually
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Inject the user and manually set the query parameters
	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	q := req.URL.Query()
	q.Add("dataset_id", dataset.ID.String())
	q.Add("group_by", "category")
	q.Add("column", "value")
	req.URL.RawQuery = q.Encode()

	// Wrap GroupDatasetBy in an ad-hoc handler for testing
	handlerFunc := func(c *gin.Context) {
		datasetIDStr := c.Query("dataset_id")
		groupBy := c.Query("group_by")
		column := c.Query("column")

		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil || groupBy == "" || column == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
			return
		}

		userID, err := handlers.GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		grouped, err := handler.GroupDatasetBy(c.Request.Context(), datasetID, userID, groupBy, column)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, grouped)
	}

	// Call the handler directly
	handlerFunc(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"A": [10, 15], "B": [5]}`, w.Body.String())
}

func TestGroupedSumHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "sumuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Sales", "GroupedSum Test")

	// Insert fields and values manually
	regionFieldID := uuid.New()
	amountFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, regionFieldID, "region", "string")
	testutils.InsertTestField(t, repo, dataset.ID, amountFieldID, "amount", "number")

	// First row
	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, regionFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, amountFieldID, "10")

	// Second row
	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, regionFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, amountFieldID, "15")

	// Third row
	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, regionFieldID, "B")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, amountFieldID, "5")

	// Set up request and context
	req := httptest.NewRequest(http.MethodGet, "/analytics/aggregation/grouped-sum?dataset_id="+dataset.ID.String()+"&group_by=region&column=amount", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Auth middleware injection (optional depending on setup)
	c = testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.GroupedSumHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"A": 25, "B": 5}`, w.Body.String())
}

func TestGroupedMeanHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "meanuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Grades", "GroupedMean Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "type", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	r1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, groupFieldID, "X")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, valueFieldID, "10")

	r2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, groupFieldID, "X")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, valueFieldID, "30")

	r3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, groupFieldID, "Y")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, valueFieldID, "20")

	req := httptest.NewRequest(http.MethodGet,
		"/analytics/aggregation/grouped-mean?dataset_id="+dataset.ID.String()+"&group_by=type&column=value",
		nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.GroupedMeanHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"X": 20, "Y": 20}`, w.Body.String())
}

func TestGroupedCountHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "countuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Counts", "GroupedCount Test")

	groupFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "group", "string")

	r1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, groupFieldID, "G1")
	r2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, groupFieldID, "G1")
	r3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, groupFieldID, "G2")

	req := httptest.NewRequest(http.MethodGet,
		"/analytics/aggregation/grouped-count?dataset_id="+dataset.ID.String()+"&group_by=group",
		nil)
	w := httptest.NewRecorder()
	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.GroupedCountHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"G1": 2, "G2": 1}`, w.Body.String())
}

func TestGroupedMinHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "minuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Scores", "GroupedMin Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "category", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "score", "number")

	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), groupFieldID, "alpha")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), valueFieldID, "9")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), groupFieldID, "alpha")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), valueFieldID, "3")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), groupFieldID, "beta")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), valueFieldID, "7")

	req := httptest.NewRequest(http.MethodGet,
		"/analytics/aggregation/grouped-min?dataset_id="+dataset.ID.String()+"&group_by=category&column=score",
		nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.GroupedMinHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"alpha": 3, "beta": 7}`, w.Body.String())
}

func TestGroupedMaxHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "maxuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Salaries", "GroupedMax Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "dept", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "salary", "number")

	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), groupFieldID, "HR")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), valueFieldID, "50000")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), groupFieldID, "HR")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), valueFieldID, "70000")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), groupFieldID, "IT")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, uuid.New(), valueFieldID, "80000")

	req := httptest.NewRequest(http.MethodGet,
		"/analytics/aggregation/grouped-max?dataset_id="+dataset.ID.String()+"&group_by=dept&column=salary",
		nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.GroupedMaxHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"HR": 70000, "IT": 80000}`, w.Body.String())
}

func TestGroupedMedianHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "medianuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Midterms", "GroupedMedian Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "group", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "score", "number")

	// Group A: 50, 70, 90 → median: 70
	// Group B: 60, 80     → median: 70 (mean of 60, 80)
	r1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, groupFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, valueFieldID, "50")

	r2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, groupFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, valueFieldID, "70")

	r3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, groupFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, valueFieldID, "90")

	r4 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r4, groupFieldID, "B")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r4, valueFieldID, "60")

	r5 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r5, groupFieldID, "B")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r5, valueFieldID, "80")

	req := httptest.NewRequest(http.MethodGet,
		"/analytics/aggregation/grouped-median?dataset_id="+dataset.ID.String()+"&group_by=group&column=score",
		nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.GroupedMedianHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"A": 70, "B": 70}`, w.Body.String())
}

func TestGroupedStdDevHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "stddevuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Deviations", "GroupedStdDev Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "group", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	// Group X: [10, 10, 10] → std dev = 0
	// Group Y: [10, 20, 30] → std dev = sqrt(100) = 10
	r1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, groupFieldID, "X")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, valueFieldID, "10")

	r2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, groupFieldID, "X")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, valueFieldID, "10")

	r3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, groupFieldID, "X")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, valueFieldID, "10")

	r4 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r4, groupFieldID, "Y")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r4, valueFieldID, "10")

	r5 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r5, groupFieldID, "Y")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r5, valueFieldID, "20")

	r6 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r6, groupFieldID, "Y")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r6, valueFieldID, "30")

	req := httptest.NewRequest(http.MethodGet,
		"/analytics/aggregation/grouped-stddev?dataset_id="+dataset.ID.String()+"&group_by=group&column=value",
		nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.GroupedStdDevHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"X": 0, "Y": 10}`, w.Body.String())
}

func setupPivotTestDataset(t *testing.T, repo *database.Repository, email string) (database.User, database.Dataset) {
	user := testutils.CreateTestUser(t, repo, email)
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Sales Dataset", "For pivot table tests")

	regionField := uuid.New()
	productField := uuid.New()
	salesField := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, regionField, "region", "string")
	testutils.InsertTestField(t, repo, dataset.ID, productField, "product", "string")
	testutils.InsertTestField(t, repo, dataset.ID, salesField, "sales", "numeric")

	testutils.InsertTestRecordMulti(t, repo, dataset.ID, map[uuid.UUID]string{
		regionField:  "North",
		productField: "A",
		salesField:   "100",
	})
	testutils.InsertTestRecordMulti(t, repo, dataset.ID, map[uuid.UUID]string{
		regionField:  "North",
		productField: "B",
		salesField:   "200",
	})
	testutils.InsertTestRecordMulti(t, repo, dataset.ID, map[uuid.UUID]string{
		regionField:  "South",
		productField: "A",
		salesField:   "300",
	})

	return user, dataset
}

func testPivotHandler(t *testing.T, route string, aggFunc string) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user, dataset := setupPivotTestDataset(t, repo, fmt.Sprintf("%s@example.com", aggFunc))

	valueFieldParam := ""
	if aggFunc != "count" {
		valueFieldParam = "&value_field=sales"
	}

	url := fmt.Sprintf("/analytics/%s?dataset_id=%s&row_field=region&column_field=product%s",
		route, dataset.ID.String(), valueFieldParam)

	req := httptest.NewRequest(http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.PivotTableHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]map[string]float64
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	// Basic structure assertions
	assert.Contains(t, result, "North")
	assert.Contains(t, result["North"], "A")
	assert.Contains(t, result["North"], "B")
	assert.Contains(t, result, "South")
	assert.Contains(t, result["South"], "A")
}

func TestPivotSumHandler(t *testing.T)    { testPivotHandler(t, "/aggregation/pivot-sum", "sum") }
func TestPivotMeanHandler(t *testing.T)   { testPivotHandler(t, "/aggregation/pivot-mean", "mean") }
func TestPivotCountHandler(t *testing.T)  { testPivotHandler(t, "/aggregation/pivot-count", "count") }
func TestPivotMinHandler(t *testing.T)    { testPivotHandler(t, "/aggregation/pivot-min", "min") }
func TestPivotMaxHandler(t *testing.T)    { testPivotHandler(t, "/aggregation/pivot-max", "max") }
func TestPivotMedianHandler(t *testing.T) { testPivotHandler(t, "/aggregation/pivot-median", "median") }
func TestPivotStdDevHandler(t *testing.T) { testPivotHandler(t, "/aggregation/pivot-stddev", "stddev") }

func TestDropRowsWithMissingHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}

	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "drop-missing@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Drop Missing", "Test drop missing")

	scoreField := uuid.New()
	ageField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, scoreField, "score", "numeric")
	testutils.InsertTestField(t, repo, dataset.ID, ageField, "age", "numeric")

	// Insert records
	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, scoreField, "85")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, ageField, "30")

	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, scoreField, "")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, ageField, "22")

	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, scoreField, "90")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, ageField, "")

	record4 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record4, scoreField, "95")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record4, ageField, "28")

	bodyJSON := `{"columns": ["score", "age"]}`
	url := "/analytics/cleaning/drop-rows-with-missing?datasetID=" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.DropRowsWithMissingHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"rows":[["85","30"],["95","28"]]}`, w.Body.String())
}

func TestFillMissingWithHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}

	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "fill-missing@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Fill Missing", "Test fill")

	scoreField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, scoreField, "score", "numeric")

	// Insert records with missing values
	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, scoreField, "")

	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, scoreField, "10")

	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, scoreField, "")

	bodyJSON := `{"column": "score", "value": 0}`
	url := "/analytics/cleaning/fill-missing-with?datasetID=" + dataset.ID.String() + "&defaultValue=0"
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.FillMissingWithHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"rows":[["0"],["10"],["0"]]}`, w.Body.String())
}

func TestApplyLogTransformationHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "log@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Log Test", "log transform test")

	valueField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, valueField, "value", "numeric")

	for _, val := range []string{"1", "10", "100"} {
		testutils.InsertTestRecord(t, repo, dataset.ID, valueField, val)
	}

	bodyJSON := `{"column":"value"}`
	url := "/analytics/cleaning/log?datasetID=" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.ApplyLogTransformationHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"rows":[["0"],["2.302585092994046"],["4.605170185988092"]]}`, w.Body.String())
}

func TestNormalizeColumnHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "normalize@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Normalize Test", "test norm")

	scoreField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, scoreField, "score", "numeric")

	for _, val := range []string{"50", "100", "150"} {
		testutils.InsertTestRecord(t, repo, dataset.ID, scoreField, val)
	}

	bodyJSON := `{"column":0,"name":"Normalized Score","description":"Normalized scores"}`
	url := "/analytics/cleaning/normalize?datasetID=" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.NormalizeColumnHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Rows [][]any `json:"rows"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expected := [][]any{
		{"50", 0.0},
		{"100", 0.5},
		{"150", 1.0},
	}
	assert.Equal(t, expected, response.Rows)
}

func TestStandardizeColumnHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "standardize@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Standardize Dataset", "test stddev")

	scoreField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, scoreField, "score", "numeric")

	for _, val := range []string{"50", "100", "150"} {
		testutils.InsertTestRecord(t, repo, dataset.ID, scoreField, val)
	}

	bodyJSON := `{"column":0,"name":"Standardized Score","description":"z-scores"}`
	url := "/analytics/cleaning/standardize-column?datasetID=" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.StandardizeColumnHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"Column standardized successfully"`)
}

func TestDropColumnsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}
	testutils.CleanDB(repo)

	email := fmt.Sprintf("testuser+%d@example.com", time.Now().UnixNano())
	user := testutils.CreateTestUser(t, repo, email)
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Test Dataset", "For column drop test")

	scoreFieldID := uuid.New()
	ageFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, scoreFieldID, "score", "numeric")
	testutils.InsertTestField(t, repo, dataset.ID, ageFieldID, "age", "numeric")

	recordID := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, scoreFieldID, "95")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, ageFieldID, "30")

	// Ensure fields exist before dropping
	params := database.GetDatasetFieldParams{
		ID:        scoreFieldID,
		DatasetID: dataset.ID,
	}
	_, err := repo.Queries.GetDatasetField(context.Background(), params)
	require.NoError(t, err)
	params.ID = ageFieldID
	_, err = repo.Queries.GetDatasetField(context.Background(), params)
	require.NoError(t, err)

	bodyJSON := fmt.Sprintf(`{"columns":["%s","%s"]}`, scoreFieldID.String(), ageFieldID.String())
	url := "/analytics/cleaning/drop-columns/" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req
	c.Params = gin.Params{{Key: "datasetID", Value: dataset.ID.String()}}

	handler.DropColumnsHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Assert fields no longer exist
	params.ID = scoreFieldID
	_, err = repo.Queries.GetDatasetField(context.Background(), params)
	require.Error(t, err)
	params.ID = ageFieldID
	_, err = repo.Queries.GetDatasetField(context.Background(), params)
	require.Error(t, err)
}

func TestRenameColumnHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}

	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "rename@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Rename Test", "rename columns test")

	// Insert a single column
	columnID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, columnID, "old_column", "string")

	// Insert a few rows
	testutils.InsertTestRecord(t, repo, dataset.ID, columnID, "alpha")
	testutils.InsertTestRecord(t, repo, dataset.ID, columnID, "beta")
	testutils.InsertTestRecord(t, repo, dataset.ID, columnID, "gamma")

	// Create rename request body
	bodyJSON := `{"new_headers": ["new_column"]}`

	url := "/analytics/cleaning/rename-columns/" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req
	c.Params = gin.Params{{Key: "datasetID", Value: dataset.ID.String()}}

	handler.RenameColumnsHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"Columns renamed successfully"`)
	assert.Contains(t, w.Body.String(), `"data":[["alpha"],["beta"],["gamma"]]`)
}

func TestPearsonHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "pearson@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Pearson Dataset", "Test")

	fieldX := uuid.New()
	fieldY := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldX, "x", "number")
	testutils.InsertTestField(t, repo, dataset.ID, fieldY, "y", "number")

	vals := [][]string{
		{"1", "2"},
		{"2", "4"},
		{"3", "6"},
	}
	for _, pair := range vals {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldX, pair[0])
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldY, pair[1])
	}

	body := `{"x_column": 0, "y_column": 1}`
	url := "/analytics/correlation/pearson-correlation?id=" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: dataset.ID.String()}}

	handler.PearsonHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"pearson":1`)
}

func TestSpearmanHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "spearman@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Spearman Dataset", "Test")

	fieldX := uuid.New()
	fieldY := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldX, "x", "number")
	testutils.InsertTestField(t, repo, dataset.ID, fieldY, "y", "number")

	vals := [][]string{
		{"1", "3"},
		{"2", "2"},
		{"3", "1"},
	}
	for _, pair := range vals {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldX, pair[0])
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldY, pair[1])
	}

	body := `{"x_column": 0, "y_column": 1}`
	url := "/analytics/correlation/spearman-correlation?id=" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: dataset.ID.String()}}

	handler.SpearmanHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"spearman":-1`)
}

func TestCorrelationMatrixHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "corrmatrix@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Correlation Matrix Dataset", "Test")

	fieldA := uuid.New()
	fieldB := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldA, "a", "number")
	testutils.InsertTestField(t, repo, dataset.ID, fieldB, "b", "number")

	vals := [][]string{
		{"1", "2"},
		{"2", "4"},
	}
	for _, pair := range vals {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldA, pair[0])
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldB, pair[1])
	}

	body := `{"columns": [0, 1], "method": "pearson"}`
	url := "/analytics/correlation/correlation-matrix?id=" + dataset.ID.String()
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: dataset.ID.String()}}

	handler.CorrelationMatrixHandler(c)

	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]map[string]float64
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	require.Contains(t, result, "a")
	require.Contains(t, result["a"], "b")
	require.InDelta(t, 1.0, result["a"]["b"], 0.05)
}

func TestMeanHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "meanuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Mean Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")

	for _, v := range []string{"10", "20", "30"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/mean?dataset_id="+dataset.ID.String()+"&column=score", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.MeanHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"mean": 20}`, w.Body.String())
}

func TestMedianHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "medianuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Median Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")

	for _, v := range []string{"10", "20", "30", "40", "50"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/median?dataset_id="+dataset.ID.String()+"&column=value", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.MedianHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"median": 30}`, w.Body.String())
}

func TestStdDevHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "stddevuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "StdDev Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")

	for _, v := range []string{"25", "30", "35"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/stddev?dataset_id="+dataset.ID.String()+"&column=value", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.StdDevHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"stddev":5`)
}

func TestVarianceHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "varuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Variance Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "height", "number")

	for _, v := range []string{"160", "170", "180"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/variance?dataset_id="+dataset.ID.String()+"&column=height", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.VarianceHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"variance":100`)
}

func TestMinHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "minuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Min Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")

	for _, v := range []string{"90", "85", "70"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/min?dataset_id="+dataset.ID.String()+"&column=score", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.MinHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"min":70`)
}

func TestMaxHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "maxuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Max Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")

	for _, v := range []string{"90", "85", "100"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/max?dataset_id="+dataset.ID.String()+"&column=score", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.MaxHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"max":100`)
}

func TestRangeHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "rangeuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Range Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "num", "number")

	for _, v := range []string{"10", "50", "30"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/range?dataset_id="+dataset.ID.String()+"&column=num", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.RangeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"range":40`)
}

func TestSumHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "sumuser2@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Sum Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "points", "number")

	for _, v := range []string{"10", "20", "30"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/sum?dataset_id="+dataset.ID.String()+"&column=points", nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.SumHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"sum":60`)
}

func TestModeHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "modeuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Colors", "Mode Test")

	colorFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, colorFieldID, "color", "string")

	// Insert values for mode calculation
	values := []string{"red", "blue", "red", "green", "red"}
	for _, v := range values {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, colorFieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/mode?dataset_id="+dataset.ID.String()+"&column=color", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c = testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.ModeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"mode":"red"`)
}

func TestCountHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "countuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "People", "Count Test")

	nameFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, nameFieldID, "name", "string")

	// Insert values for counting
	names := []string{"Alice", "Bob", "Alice"}
	for _, v := range names {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, nameFieldID, v)
	}

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/count?dataset_id="+dataset.ID.String()+"&column=name", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c = testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.CountHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"count":3`)
}

func TestHistogramHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "histouser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Values", "Histogram Test")

	valueFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	values := []string{"1", "2", "3", "4", "5", "6"}
	for _, val := range values {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, valueFieldID, val)
	}

	url := fmt.Sprintf("/analytics/distribution/histogram?dataset_id=%s&column=value&num_bins=3", dataset.ID.String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c = testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.HistogramHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"labels"`)
	assert.Contains(t, w.Body.String(), `"counts"`)
}

func TestKDEHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "kdeuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Values", "KDE Test")

	valueFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	values := []string{"1", "2", "3", "4", "5"}
	for _, val := range values {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, valueFieldID, val)
	}

	url := fmt.Sprintf("/analytics/distribution/kde?dataset_id=%s&column=value&num_points=10&bandwidth=1.0", dataset.ID.String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c = testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.KDEHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"labels"`)
	assert.Contains(t, w.Body.String(), `"densities"`)
}

func TestFilterSortHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.DatasetHandler{Service: service}

	user := testutils.CreateTestUser(t, repo, "filtersortuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "People", "FilterSort Test")

	nameFieldID := uuid.New()
	ageFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, nameFieldID, "name", "string")
	testutils.InsertTestField(t, repo, dataset.ID, ageFieldID, "age", "number")

	// Alice (age 24)
	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, nameFieldID, "Alice")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, ageFieldID, "24")

	// Bob (age 30)
	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, nameFieldID, "Bob")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, ageFieldID, "30")

	// Charlie (age 35)
	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, nameFieldID, "Charlie")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, ageFieldID, "35")

	query := fmt.Sprintf(
		"/analytics/filtersort/filter-sort?dataset_id=%s&filter_col=age&filter_op=gt&filter_val=25&sort_by=age&order=asc",
		dataset.ID.String(),
	)

	req := httptest.NewRequest(http.MethodGet, query, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	c = testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.FilterSortHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"name":"Bob","age":"30"}`)
	assert.Contains(t, w.Body.String(), `{"name":"Charlie","age":"35"}`)
}

func TestZScoreOutliersHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{DatasetService: service}

	user := testutils.CreateTestUser(t, repo, "zscore@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "ZScore Dataset", "Z-score outlier test")

	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")

	values := []string{"10", "15", "100", "20"}
	for _, val := range values {
		rid := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, rid, fieldID, val)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/analytics/outliers/zscore?dataset_id=%s&column=value&threshold=1.4", dataset.ID.String()), nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.ZScoreOutliersHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"indices":[2]`)
}

func TestIQROutliersHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{DatasetService: service}

	user := testutils.CreateTestUser(t, repo, "iqr@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "IQR Dataset", "IQR outlier test")

	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")

	values := []string{"10", "15", "25", "100", "30"}
	for _, val := range values {
		rid := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, rid, fieldID, val)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/analytics/outliers/iqr?dataset_id=%s&column=value", dataset.ID.String()), nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.IQROutliersHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"indices":`)
}

func TestBoxPlotHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{DatasetService: service}

	user := testutils.CreateTestUser(t, repo, "boxplot@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "BoxPlot Dataset", "BoxPlot test")

	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")

	values := []string{"10", "15", "25", "100", "30"}
	for _, val := range values {
		rid := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, rid, fieldID, val)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/analytics/distribution/boxplot?dataset_id=%s&column=value", dataset.ID.String()), nil)
	w := httptest.NewRecorder()

	c := testutils.GetAuthenticatedContext(user.ID, user.Email)
	c.Request = req

	handler.BoxPlotHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"stats"`)
}
