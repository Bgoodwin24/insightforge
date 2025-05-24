package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
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

	user := testutils.CreateTestUser(t, repo, "groupbyuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Sales", "GroupBy Test")

	categoryFieldID := uuid.New()
	valueFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, categoryFieldID, "category", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, categoryFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, valueFieldID, "10")

	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, categoryFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, valueFieldID, "15")

	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, categoryFieldID, "B")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, valueFieldID, "5")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	q := req.URL.Query()
	q.Add("dataset_id", dataset.ID.String())
	q.Add("group_by", "category")
	q.Add("column", "value")
	req.URL.RawQuery = q.Encode()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID.String())
	c.Set("userEmail", user.Email)

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

	handlerFunc(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"A": [10, 15], "B": [5]}`, w.Body.String())
}

func TestGroupedSumHandler(t *testing.T) {
	// Setup
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        datasetService,
		DatasetService: datasetService,
	}

	// Create user and dataset
	user := testutils.CreateTestUser(t, repo, "sumuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Sales", "GroupedSum Test")

	// Insert fields and values manually
	regionFieldID := uuid.New()
	amountFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, regionFieldID, "region", "string")
	testutils.InsertTestField(t, repo, dataset.ID, amountFieldID, "amount", "number")

	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, regionFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, amountFieldID, "10")

	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, regionFieldID, "A")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, amountFieldID, "15")

	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, regionFieldID, "B")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, amountFieldID, "5")

	// Create JWT token
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup Gin router with real middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	analyticsGroup := router.Group("/analytics")
	analyticsGroup.GET("/aggregation/grouped-sum", handler.GroupedSumHandler)

	// Build authenticated request
	url := fmt.Sprintf("/analytics/aggregation/grouped-sum?dataset_id=%s&group_by=region&column=amount", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Host = "localhost"
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})
	w := httptest.NewRecorder()

	// Send request through router
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"results": {"A": 25, "B": 5}}`, w.Body.String())
}

func TestGroupedMeanHandler(t *testing.T) {
	// Setup
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        datasetService,
		DatasetService: datasetService,
	}

	// Create user and dataset
	user := testutils.CreateTestUser(t, repo, "meanuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Grades", "GroupedMean Test")

	// Insert fields and values manually
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

	// Generate token
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	analyticsGroup := router.Group("/analytics")
	analyticsGroup.GET("/aggregation/grouped-mean", handler.GroupedMeanHandler)

	// Build request
	url := fmt.Sprintf("/analytics/aggregation/grouped-mean?dataset_id=%s&group_by=type&column=value", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Host = "localhost"
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"results": {"X": 20, "Y": 20}}`, w.Body.String())
}

func TestGroupedCountHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{SecretKey: jwtSecret, TokenDuration: 24 * time.Hour}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: datasetService, DatasetService: datasetService}

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

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/analytics/aggregation/grouped-count", handler.GroupedCountHandler)

	url := fmt.Sprintf("/analytics/aggregation/grouped-count?dataset_id=%s&group_by=group", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"results": {"G1": 2, "G2": 1}}`, w.Body.String())
}

func TestGroupedMinHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{SecretKey: jwtSecret, TokenDuration: 24 * time.Hour}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: datasetService, DatasetService: datasetService}

	user := testutils.CreateTestUser(t, repo, "minuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Scores", "GroupedMin Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "category", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "score", "number")

	// alpha: 9, 3 → min: 3
	// beta: 7 → min: 7
	// alpha: 9, 3 → min: 3
	// beta: 7 → min: 7

	// First record for alpha
	r1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, groupFieldID, "alpha")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, valueFieldID, "9")

	// Second record for alpha
	r2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, groupFieldID, "alpha")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, valueFieldID, "3")

	// Record for beta
	r3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, groupFieldID, "beta")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, valueFieldID, "7")

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/analytics/aggregation/grouped-min", handler.GroupedMinHandler)

	url := fmt.Sprintf("/analytics/aggregation/grouped-min?dataset_id=%s&group_by=category&column=score", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"results": {"alpha": 3, "beta": 7}}`, w.Body.String())
}

func TestGroupedMaxHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{SecretKey: jwtSecret, TokenDuration: 24 * time.Hour}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: datasetService, DatasetService: datasetService}

	user := testutils.CreateTestUser(t, repo, "maxuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Salaries", "GroupedMax Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "dept", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "salary", "number")

	// HR: 50000, 70000 → max: 70000
	// IT: 80000 → max: 80000

	// First HR record
	r1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, groupFieldID, "HR")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r1, valueFieldID, "50000")

	// Second HR record
	r2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, groupFieldID, "HR")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r2, valueFieldID, "70000")

	// IT record
	r3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, groupFieldID, "IT")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, r3, valueFieldID, "80000")

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/analytics/aggregation/grouped-max", handler.GroupedMaxHandler)

	url := fmt.Sprintf("/analytics/aggregation/grouped-max?dataset_id=%s&group_by=dept&column=salary", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"results": {"HR": 70000, "IT": 80000}}`, w.Body.String())
}

func TestGroupedMedianHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{SecretKey: jwtSecret, TokenDuration: 24 * time.Hour}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: datasetService, DatasetService: datasetService}

	user := testutils.CreateTestUser(t, repo, "medianuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Midterms", "GroupedMedian Test")

	groupFieldID := uuid.New()
	valueFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, groupFieldID, "group", "string")
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "score", "number")

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

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/analytics/aggregation/grouped-median", handler.GroupedMedianHandler)

	url := fmt.Sprintf("/analytics/aggregation/grouped-median?dataset_id=%s&group_by=group&column=score", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"results": {"A": 70, "B": 70}}`, w.Body.String())
}

func TestGroupedStdDevHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{SecretKey: jwtSecret, TokenDuration: 24 * time.Hour}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: datasetService, DatasetService: datasetService}

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

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/analytics/aggregation/grouped-stddev", handler.GroupedStdDevHandler)

	url := fmt.Sprintf("/analytics/aggregation/grouped-stddev?dataset_id=%s&group_by=group&column=value", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"results": {"X": 0, "Y": 10}}`, w.Body.String())
}

func setupPivotTestDataset(t *testing.T, repo *database.Repository, email string) (
	database.User, database.Dataset, string, string, string,
) {
	user := testutils.CreateTestUser(t, repo, email)
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Sales Dataset", "For pivot table tests")

	regionField := uuid.New()
	productField := uuid.New()
	salesField := uuid.New()

	regionName := "region"
	productName := "product"
	salesName := "sales"

	testutils.InsertTestField(t, repo, dataset.ID, regionField, regionName, "string")
	testutils.InsertTestField(t, repo, dataset.ID, productField, productName, "string")
	testutils.InsertTestField(t, repo, dataset.ID, salesField, salesName, "numeric")

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

	return user, dataset, regionName, productName, salesName
}

func testPivotHandler(t *testing.T, route string, aggFunc string) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service}

	user, dataset, regionName, productName, salesName := setupPivotTestDataset(t, repo, fmt.Sprintf("%s@example.com", aggFunc))

	url := fmt.Sprintf(
		"/analytics/aggregation/%s?dataset_id=%s&row_field=%s&column=%s",
		route,
		dataset.ID.String(),
		regionName,
		productName,
	)

	if aggFunc != "count" {
		url += fmt.Sprintf("&value_field=%s", salesName)
	}

	url += fmt.Sprintf("&agg_func=%s", aggFunc)
	log.Println(url)

	req := httptest.NewRequest(http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", user.ID.String())
	c.Set("user_email", user.Email)

	handler.PivotTableHandler(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 but got %d, body: %s", w.Code, w.Body.String())
	}

	type PivotResponse struct {
		AggFunc    string                        `json:"agg_func"`
		Column     string                        `json:"column"`
		RowField   string                        `json:"row_field"`
		ValueField string                        `json:"value_field"`
		Results    map[string]map[string]float64 `json:"results"`
	}

	var resp PivotResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Failed to unmarshal response")

	assert.Contains(t, resp.Results, "North")
	assert.Contains(t, resp.Results["North"], "A")
	assert.Contains(t, resp.Results["North"], "B")
	assert.Contains(t, resp.Results, "South")
	assert.Contains(t, resp.Results["South"], "A")
}

func TestPivotSumHandler(t *testing.T)    { testPivotHandler(t, "pivot-sum", "sum") }
func TestPivotMeanHandler(t *testing.T)   { testPivotHandler(t, "pivot-mean", "mean") }
func TestPivotCountHandler(t *testing.T)  { testPivotHandler(t, "pivot-count", "count") }
func TestPivotMinHandler(t *testing.T)    { testPivotHandler(t, "pivot-min", "min") }
func TestPivotMaxHandler(t *testing.T)    { testPivotHandler(t, "pivot-max", "max") }
func TestPivotMedianHandler(t *testing.T) { testPivotHandler(t, "pivot-median", "median") }
func TestPivotStdDevHandler(t *testing.T) { testPivotHandler(t, "pivot-stddev", "stddev") }

func TestDropRowsWithMissingHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := handlers.NewDatasetHandler(datasetService)

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

	// Generate JWT
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))

	analyticsGroup := router.Group("/analytics")
	analyticsGroup.POST("/cleaning/drop-rows-with-missing", handler.DropRowsWithMissingHandler)

	// Prepare request
	bodyJSON := `{"columns": ["score", "age"]}`
	url := fmt.Sprintf("/analytics/cleaning/drop-rows-with-missing?dataset_id=%s", dataset.ID.String())

	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Validate response
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"rows":[["score","age"],["85","30"],["95","28"]]}`, w.Body.String())
}

func TestFillMissingWithHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	service := services.NewDatasetService(repo)
	handler := handlers.NewDatasetHandler(service)

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

	// Generate JWT
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup Gin router with auth
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))

	analyticsGroup := router.Group("/analytics")
	analyticsGroup.POST("/cleaning/fill-missing-with", handler.FillMissingWithHandler)

	// Prepare request WITHOUT JSON body
	url := fmt.Sprintf("/analytics/cleaning/fill-missing-with?dataset_id=%s&defaultValue=0", dataset.ID.String())

	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"rows":[["0"],["10"],["0"]]}`, w.Body.String())
}

func TestApplyLogTransformationHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := handlers.NewDatasetHandler(service)
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "log@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Log Test", "log transform test")

	valueField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, valueField, "value", "numeric")

	for _, val := range []string{"1", "10", "100"} {
		testutils.InsertTestRecord(t, repo, dataset.ID, valueField, val)
	}

	// Setup JWT for auth middleware if needed (similar to your fill-missing test)
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup Gin router with your handler and auth middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))

	analyticsGroup := router.Group("/analytics")
	analyticsGroup.POST("/cleaning/log", handler.ApplyLogTransformationHandler)

	url := fmt.Sprintf("/analytics/cleaning/log?dataset_id=%s&col=0", dataset.ID.String())
	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"rows":[["0"],["2.302585092994046"],["4.605170185988092"]]}`, w.Body.String())
}

func TestNormalizeColumnHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := handlers.NewDatasetHandler(service)
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "normalize@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Normalize Test", "test norm")

	scoreField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, scoreField, "score", "numeric")

	for _, val := range []string{"50", "100", "150"} {
		testutils.InsertTestRecord(t, repo, dataset.ID, scoreField, val)
	}

	// Set up JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up Gin router with auth
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))

	router.POST("/analytics/cleaning/normalize", handler.NormalizeColumnHandler)

	bodyJSON := `{"column":0,"name":"Normalized Score","description":"Normalized scores"}`
	url := fmt.Sprintf("/analytics/cleaning/normalize?dataset_id=%s", dataset.ID.String())
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Rows [][]any `json:"rows"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

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
	handler := handlers.NewDatasetHandler(service)
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "standardize@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Standardize Dataset", "test stddev")

	scoreField := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, scoreField, "score", "numeric")

	for _, val := range []string{"50", "100", "150"} {
		testutils.InsertTestRecord(t, repo, dataset.ID, scoreField, val)
	}

	// Set up JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up Gin router with auth middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.POST("/analytics/cleaning/standardize-column", handler.StandardizeColumnHandler)

	bodyJSON := `{"column":0,"name":"Standardized Score","description":"z-scores"}`
	url := fmt.Sprintf("/analytics/cleaning/standardize-column?dataset_id=%s", dataset.ID.String())
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string][]float64
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	expectedValues := []float64{-1.224744871391589, 0, 1.224744871391589}
	assert.InDeltaSlice(t, expectedValues, response["Standardized Score"], 1e-9)
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
	c.Params = gin.Params{{Key: "dataset_id", Value: dataset.ID.String()}}

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
	handler := handlers.NewDatasetHandler(service)
	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "rename@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Rename Test", "rename columns test")

	// Insert a single column
	columnID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, columnID, "old_column", "string")

	// Insert rows
	testutils.InsertTestRecord(t, repo, dataset.ID, columnID, "alpha")
	testutils.InsertTestRecord(t, repo, dataset.ID, columnID, "beta")
	testutils.InsertTestRecord(t, repo, dataset.ID, columnID, "gamma")

	// Generate JWT token
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup Gin router with auth middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.POST("/analytics/cleaning/rename-columns/:dataset_id", handler.RenameColumnsHandler)

	bodyJSON := `{"new_headers": ["new_column"]}`
	url := fmt.Sprintf("/analytics/cleaning/rename-columns/%s", dataset.ID.String())
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"Columns renamed successfully"`)
	assert.Contains(t, w.Body.String(), `"data":[["new_column"],["beta"],["gamma"]]`)
}

func TestPearsonHandler(t *testing.T) {
	// Setup JWT and services
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)

	handler := handlers.NewDatasetHandler(datasetService)

	// Create test user and dataset
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

	// Generate token
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))

	analyticsGroup := router.Group("/analytics")
	analyticsGroup.GET("/correlation/pearson-correlation", handler.PearsonHandler)

	// Prepare request
	url := fmt.Sprintf("/analytics/correlation/pearson-correlation?dataset_id=%s&row_field=x&column=y", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"pearson":1`)
}

func TestSpearmanHandler(t *testing.T) {
	// Setup JWT and services
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := handlers.NewDatasetHandler(datasetService)

	testutils.CleanDB(repo)

	// Create test user and dataset
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

	// Generate JWT token
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up Gin router with middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))

	analyticsGroup := router.Group("/analytics")
	analyticsGroup.GET("/correlation/spearman-correlation", handler.SpearmanHandler)

	// Prepare GET request with query parameters
	url := fmt.Sprintf("/analytics/correlation/spearman-correlation?dataset_id=%s&row_field=x&column=y", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"spearman":-1`)
}

func TestCorrelationMatrixHandler(t *testing.T) {
	// Setup JWT and services
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := handlers.NewDatasetHandler(datasetService)

	testutils.CleanDB(repo)

	// Create user and dataset
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

	// Generate JWT token
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up router with middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))

	analyticsGroup := router.Group("/analytics")
	analyticsGroup.POST("/correlation/correlation-matrix", handler.CorrelationMatrixHandler)

	// Create and send request
	url := fmt.Sprintf("/analytics/correlation/correlation-matrix?dataset_id=%s&method=pearson&column=a&column=b", dataset.ID)
	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	require.Equal(t, http.StatusOK, w.Code)

	// Unmarshal into the correct wrapper struct
	var outer struct {
		Results map[string]map[string]float64 `json:"results"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &outer)
	require.NoError(t, err)

	require.Contains(t, outer.Results, "a")
	require.Contains(t, outer.Results["a"], "b")
	require.InDelta(t, 1.0, outer.Results["a"]["b"], 0.05)
}

func TestMeanHandler(t *testing.T) {
	// Setup
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)

	handler := &handlers.AnalyticsHandler{
		Service:        datasetService,
		DatasetService: datasetService,
	}

	// Create user and dataset
	user := testutils.CreateTestUser(t, repo, "meanuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Mean Test")

	// Insert fields and values
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")

	for _, v := range []string{"10", "20", "30"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	// Generate JWT token
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	analyticsGroup := router.Group("/analytics")
	analyticsGroup.GET("/descriptives/mean", handler.MeanHandler)

	// Authenticated request
	url := fmt.Sprintf("/analytics/descriptives/mean?dataset_id=%s&column=score", dataset.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Host = "localhost"
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"mean": 20}`, w.Body.String())
}

func TestMedianHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)

	handler := &handlers.AnalyticsHandler{
		Service:        datasetService,
		DatasetService: datasetService,
	}

	user := testutils.CreateTestUser(t, repo, "medianuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Median Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")
	for _, v := range []string{"10", "20", "30", "40", "50"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	analyticsGroup := router.Group("/analytics")
	analyticsGroup.GET("/descriptives/median", handler.MedianHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/median?dataset_id="+dataset.ID.String()+"&column=value", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"median": 30}`, w.Body.String())
}

func TestStdDevHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        datasetService,
		DatasetService: datasetService,
	}

	user := testutils.CreateTestUser(t, repo, "stddevuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "StdDev Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "value", "number")
	for _, v := range []string{"25", "30", "35"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/stddev", handler.StdDevHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/stddev?dataset_id="+dataset.ID.String()+"&column=value", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"stddev":5}`, w.Body.String())
}

func TestVarianceHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        datasetService,
		DatasetService: datasetService,
	}

	user := testutils.CreateTestUser(t, repo, "varuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Stats Dataset", "Variance Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "height", "number")
	for _, v := range []string{"160", "170", "180"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/variance", handler.VarianceHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/variance?dataset_id="+dataset.ID.String()+"&column=height", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"variance":100}`, w.Body.String())
}

func TestMinHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        service,
		DatasetService: service,
	}

	user := testutils.CreateTestUser(t, repo, "minuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Descriptive Stats", "Min Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")
	for _, v := range []string{"30", "20", "50"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, _ := jwtManager.Generate(user.ID, user.Email)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/min", handler.MinHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/min?dataset_id="+dataset.ID.String()+"&column=score", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"min":20}`, w.Body.String())
}

func TestMaxHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        service,
		DatasetService: service,
	}

	user := testutils.CreateTestUser(t, repo, "maxuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Descriptive Stats", "Max Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")
	for _, v := range []string{"10", "60", "40"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, _ := jwtManager.Generate(user.ID, user.Email)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/max", handler.MaxHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/max?dataset_id="+dataset.ID.String()+"&column=score", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"max":60}`, w.Body.String())
}

func TestRangeHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        service,
		DatasetService: service,
	}

	user := testutils.CreateTestUser(t, repo, "rangeuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Descriptive Stats", "Range Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")
	for _, v := range []string{"10", "25", "40"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, _ := jwtManager.Generate(user.ID, user.Email)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/range", handler.RangeHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/range?dataset_id="+dataset.ID.String()+"&column=score", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"range":30}`, w.Body.String())
}

func TestSumHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        service,
		DatasetService: service,
	}

	user := testutils.CreateTestUser(t, repo, "sumuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Descriptive Stats", "Sum Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "points", "number")
	for _, v := range []string{"5", "15", "30"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, _ := jwtManager.Generate(user.ID, user.Email)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/sum", handler.SumHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/sum?dataset_id="+dataset.ID.String()+"&column=points", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"sum":50}`, w.Body.String())
}

func TestModeHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        service,
		DatasetService: service,
	}

	user := testutils.CreateTestUser(t, repo, "modeuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Descriptive Stats", "Mode Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "score", "number")
	for _, v := range []string{"100", "100", "90"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, _ := jwtManager.Generate(user.ID, user.Email)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/mode", handler.ModeHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/mode?dataset_id="+dataset.ID.String()+"&column=score", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"mode":[100]}`, w.Body.String())
}

func TestCountHandler(t *testing.T) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{
		Service:        service,
		DatasetService: service,
	}

	user := testutils.CreateTestUser(t, repo, "countuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Descriptive Stats", "Count Test")
	fieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, fieldID, "visits", "number")
	for _, v := range []string{"1", "1", "1", "1"} {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, fieldID, v)
	}

	token, _ := jwtManager.Generate(user.ID, user.Email)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics")
	group.GET("/descriptives/count", handler.CountHandler)

	req := httptest.NewRequest(http.MethodGet, "/analytics/descriptives/count?dataset_id="+dataset.ID.String()+"&column=visits", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"count":4}`, w.Body.String())
}

func TestHistogramHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service, DatasetService: service}

	user := testutils.CreateTestUser(t, repo, "histouser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Values", "Histogram Test")

	valueFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	values := []string{"1", "2", "3", "4", "5", "6"}
	for _, val := range values {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, valueFieldID, val)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics/distribution")
	group.GET("/histogram", handler.HistogramHandler)

	url := fmt.Sprintf("/analytics/distribution/histogram?dataset_id=%s&column=value&num_bins=3", dataset.ID.String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"labels"`)
	assert.Contains(t, w.Body.String(), `"counts"`)
}

func TestKDEHandler(t *testing.T) {
	repo := testutils.SetupTestRepo()
	service := services.NewDatasetService(repo)
	handler := &handlers.AnalyticsHandler{Service: service, DatasetService: service}

	user := testutils.CreateTestUser(t, repo, "kdeuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "Values", "KDE Test")

	valueFieldID := uuid.New()
	testutils.InsertTestField(t, repo, dataset.ID, valueFieldID, "value", "number")

	values := []string{"1", "2", "3", "4", "5"}
	for _, val := range values {
		recordID := uuid.New()
		testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, recordID, valueFieldID, val)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}

	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics/distribution")
	group.GET("/kde", handler.KDEHandler)

	url := fmt.Sprintf("/analytics/distribution/kde?dataset_id=%s&column=value&num_points=10&bandwidth=1.0", dataset.ID.String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"labels"`)
	assert.Contains(t, w.Body.String(), `"densities"`)
}

func TestFilterSortHandler(t *testing.T) {
	// Setup JWT manager
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
	}

	db := testutils.SetupDB()
	repo := database.NewRepository(db)
	datasetService := services.NewDatasetService(repo)
	handler := handlers.NewDatasetHandler(datasetService)

	testutils.CleanDB(repo)

	user := testutils.CreateTestUser(t, repo, "filtersortuser@example.com")
	dataset := testutils.CreateTestDataset(t, repo, user.ID, "People", "FilterSort Test")

	nameFieldID := uuid.New()
	ageFieldID := uuid.New()

	testutils.InsertTestField(t, repo, dataset.ID, nameFieldID, "name", "string")
	testutils.InsertTestField(t, repo, dataset.ID, ageFieldID, "age", "number")

	record1 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, nameFieldID, "Alice")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record1, ageFieldID, "24")

	record2 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, nameFieldID, "Bob")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record2, ageFieldID, "30")

	record3 := uuid.New()
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, nameFieldID, "Charlie")
	testutils.InsertTestValueWithRecordID(t, repo, dataset.ID, record3, ageFieldID, "35")

	// Generate JWT token
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Setup router with auth middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	analyticsGroup := router.Group("/analytics")
	analyticsGroup.GET("/filtersort/filter-sort", handler.FilterSortHandler)

	url := fmt.Sprintf("/analytics/filtersort/filter-sort?dataset_id=%s&filter_col=age&filter_op=gt&filter_val=25&sort_by=age&order=asc", dataset.ID.String())

	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// 🔄 New: Parse JSON response
	var resp struct {
		Data []map[string]string `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// 🔍 Assert the filtered & sorted result
	expected := []map[string]string{
		{"name": "Bob", "age": "30"},
		{"name": "Charlie", "age": "35"},
	}

	assert.Equal(t, expected, resp.Data)
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

	// Generate token
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up router
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics/outliers")
	group.GET("/zscore-outliers", handler.ZScoreOutliersHandler)

	url := fmt.Sprintf("/analytics/outliers/zscore-outliers?dataset_id=%s&column=value&threshold=1.4", dataset.ID.String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

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

	// Generate token
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up router and route
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics/outliers")
	group.GET("/iqr-outliers", handler.IQROutliersHandler)

	// Create request
	url := fmt.Sprintf("/analytics/outliers/iqr-outliers?dataset_id=%s&column=value", dataset.ID.String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

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

	// Generate token
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: 24 * time.Hour,
	}
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Set up router and route
	router := gin.New()
	router.Use(auth.AuthMiddleware(jwtManager))
	group := router.Group("/analytics/distribution")
	group.GET("/boxplot", handler.BoxPlotHandler)

	// Create request
	url := fmt.Sprintf("/analytics/distribution/boxplot?dataset_id=%s&column=value", dataset.ID.String())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"stats"`)
	assert.Contains(t, w.Body.String(), `"values"`)
	assert.Contains(t, w.Body.String(), `"labels"`)
}
