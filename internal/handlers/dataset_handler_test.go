package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/Bgoodwin24/insightforge/internal/handlers"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDataset_Success(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Minute*15)
	db := testutils.SetupTestRepo()
	testutils.CleanDB(db)
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	user := testutils.CreateTestUser(t, db, "createdataset@example.com")

	token, err := jwtManager.Generate(user.ID, user.Email)
	assert.NoError(t, err)

	router := gin.Default()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.POST("/datasets", handler.CreateDataset)

	body := `{"name": "My Dataset", "description": "Sample desc"}`
	req, _ := http.NewRequest(http.MethodPost, "/datasets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateDataset_Validation(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	db := testutils.SetupTestRepo()
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	user := testutils.CreateTestUser(t, db, "createdatasetfail@example.com")

	router := gin.Default()
	router.POST("/datasets", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.CreateDataset(c)
	})

	body := `{"description": "Missing name"}`
	req, _ := http.NewRequest(http.MethodPost, "/datasets", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListDatasets(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	db := testutils.SetupTestRepo()
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	user := testutils.CreateTestUser(t, db, "listdatasets@example.com")
	testutils.CreateTestDataset(t, db, user.ID, "Dataset 1", "desc")

	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Minute*15)
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	router := gin.Default()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/datasets", handler.ListDatasets)

	req, _ := http.NewRequest(http.MethodGet, "/datasets?limit=5", nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestSearchDatasets(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	db := testutils.SetupTestRepo()
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	user := testutils.CreateTestUser(t, db, "searchdatasets@example.com")
	testutils.CreateTestDataset(t, db, user.ID, "MatchMe", "desc")
	testutils.CreateTestDataset(t, db, user.ID, "NoMatch", "desc")

	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Minute*15)
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	router := gin.Default()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/datasets/search", handler.SearchDataSets)

	req, _ := http.NewRequest(http.MethodGet, "/datasets/search?search=Match", nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})

	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestGetDatasetByID(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	db := testutils.SetupTestRepo()
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	user := testutils.CreateTestUser(t, db, "getdataset@example.com")
	dataset := testutils.CreateTestDataset(t, db, user.ID, "Dataset", "desc")

	// Set up the JWT Manager and generate the token
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Minute*15)
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Create the router and apply the AuthMiddleware
	router := gin.Default()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.GET("/datasets/:id", handler.GetDatasetByID)

	// Create the request with the authorization header
	req, _ := http.NewRequest(http.MethodGet, "/datasets/"+dataset.ID.String(), nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})

	resp := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(resp, req)

	// Assert the response code
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestDeleteDatasetByID(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	db := testutils.SetupTestRepo()
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	user := testutils.CreateTestUser(t, db, "deletedataset@example.com")
	dataset := testutils.CreateTestDataset(t, db, user.ID, "Delete Me", "desc")

	// Set up the JWT Manager and generate the token
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Minute*15)
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	// Create the router and apply the AuthMiddleware
	router := gin.Default()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.DELETE("/datasets/:id", handler.DeleteDatasetsByID)

	// Create the request with the authorization header
	req, _ := http.NewRequest(http.MethodDelete, "/datasets/"+dataset.ID.String(), nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})

	resp := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(resp, req)

	// Assert the response code
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestUpdateDataset_Success(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	db := testutils.SetupTestRepo()
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	user := testutils.CreateTestUser(t, db, "updatedataset@example.com")
	dataset := testutils.CreateTestDataset(t, db, user.ID, "Old Name", "Old Description")

	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Minute*15)
	token, err := jwtManager.Generate(user.ID, user.Email)
	require.NoError(t, err)

	router := gin.Default()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.PUT("/datasets/:id", handler.UpdateDataset)

	payload := map[string]string{
		"name":        "Updated Name",
		"description": "Updated Description",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, "/datasets/"+dataset.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &updatedResp)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedResp["name"])
	desc, ok := updatedResp["description"].(string)
	require.True(t, ok)
	assert.Equal(t, "Updated Description", desc)
}

func TestUpdateDataset_Forbidden(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	db := testutils.SetupTestRepo()
	service := services.NewDatasetService(db)
	handler := handlers.NewDatasetHandler(service)

	// Create owner and intruder users
	ownerEmail := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	intruderEmail := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	// Create users in the database
	owner := testutils.CreateTestUser(t, db, ownerEmail)
	intruder := testutils.CreateTestUser(t, db, intruderEmail)

	// Create dataset for the owner
	dataset := testutils.CreateTestDataset(t, db, owner.ID, "Secret", "Don't touch this")

	// Generate a JWT token for the intruder
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Minute*15)
	token, err := jwtManager.Generate(intruder.ID, intruder.Email)
	require.NoError(t, err)

	// Set up the router with the auth middleware
	router := gin.Default()
	router.Use(auth.AuthMiddleware(jwtManager))
	router.PUT("/datasets/:id", handler.UpdateDataset)

	// Prepare the payload for the update
	payload := map[string]string{
		"name":        "Hack Attempt",
		"description": "Hacked",
	}
	body, _ := json.Marshal(payload)

	// Make the HTTP PUT request
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/datasets/%s", dataset.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})

	// Record the response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert that the response status is Forbidden (403)
	assert.Equal(t, http.StatusForbidden, resp.Code)

	// Optionally: Add assertions to check that the dataset has not been modified
	updatedDataset, err := db.Queries.GetDatasetByID(context.Background(), dataset.ID)
	require.NoError(t, err)
	assert.Equal(t, dataset.Name, updatedDataset.Name) // Verify name wasn't changed
}
