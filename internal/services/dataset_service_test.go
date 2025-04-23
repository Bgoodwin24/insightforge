package services_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/internal/testutils"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDataset(t *testing.T) {
	db := setupDB()
	defer db.Close()
	repo := database.NewRepository(db)
	cleanDB(repo)
	logger.Init()

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())

	svc := services.NewDatasetService(repo)
	user := testutils.CreateTestUser(t, repo, email)

	dataset, err := svc.CreateDataset(context.Background(), user.ID, "My Dataset", "This is a test")
	require.NoError(t, err)

	assert.Equal(t, user.ID, dataset.UserID)
	assert.Equal(t, "My Dataset", dataset.Name)
	assert.True(t, dataset.CreatedAt.Before(time.Now()))
}

func TestGetDatasetByIDForUser(t *testing.T) {
	db := setupDB()
	defer db.Close()
	repo := database.NewRepository(db)
	cleanDB(repo)
	logger.Init()

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())

	svc := services.NewDatasetService(repo)
	user := testutils.CreateTestUser(t, repo, email)

	created, _ := svc.CreateDataset(context.Background(), user.ID, "Dataset", "Get Test")

	// Should succeed
	result, err := svc.GetDatasetByIDForUser(context.Background(), user.ID, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)

	// Should fail with unauthorized user
	other := testutils.CreateTestUser(t, repo, "unauth@example.com")
	_, err = svc.GetDatasetByIDForUser(context.Background(), other.ID, created.ID)
	assert.Error(t, err)
}

func TestListDatasetsForUser(t *testing.T) {
	db := setupDB()
	defer db.Close()
	repo := database.NewRepository(db)
	cleanDB(repo)
	logger.Init()
	datasetServ := services.NewDatasetService(repo)

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())

	svc := services.NewDatasetService(repo)
	user := testutils.CreateTestUser(t, repo, email)

	// Create 3 datasets
	var firstDataset database.Dataset
	for i := 0; i < 3; i++ {
		ds, err := svc.CreateDataset(context.Background(), user.ID, fmt.Sprintf("Dataset %d", i+1), "")
		require.NoError(t, err)
		if i == 0 {
			firstDataset = ds
		}
	}

	// Add logging here to confirm creation
	datasets, _ := datasetServ.GetDatasetByIDForUser(context.Background(), user.ID, firstDataset.ID)
	fmt.Printf("Datasets in DB: %v\n", datasets)

	// Test case 1: Retrieve all datasets with limit large enough
	list, err := svc.ListDatasetsForUser(context.Background(), user.ID, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1) // Ensure the list is not empty

	// Test case 2: Retrieve first 2 datasets with pagination (limit=2, offset=0)
	list, err = svc.ListDatasetsForUser(context.Background(), user.ID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, list, 2) // We expect 2 datasets in the first page

	// Test case 3: Retrieve the next 1 dataset with pagination (limit=2, offset=2)
	list, err = svc.ListDatasetsForUser(context.Background(), user.ID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, list, 1) // We expect 1 dataset on the second page
}

func TestUpdateDataset(t *testing.T) {
	db := setupDB()
	defer db.Close()
	repo := database.NewRepository(db)
	cleanDB(repo)
	logger.Init()

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())

	svc := services.NewDatasetService(repo)
	user := testutils.CreateTestUser(t, repo, email)

	dataset, _ := svc.CreateDataset(context.Background(), user.ID, "Old Name", "")
	updated, err := svc.UpdateDataset(context.Background(), dataset.ID, "New Name", "Updated")
	require.NoError(t, err)

	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "Updated", updated.Description.String)
}

func TestDeleteDataset(t *testing.T) {
	db := setupDB()
	defer db.Close()
	repo := database.NewRepository(db)
	cleanDB(repo)
	logger.Init()

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())

	svc := services.NewDatasetService(repo)
	user := testutils.CreateTestUser(t, repo, email)

	dataset, _ := svc.CreateDataset(context.Background(), user.ID, "To Delete", "")
	err := svc.DeleteDataset(context.Background(), dataset.ID, user.ID)
	assert.NoError(t, err)

	_, err = repo.Queries.GetDatasetByID(context.Background(), dataset.ID)
	assert.Error(t, err) // Should not exist
}

func TestSearchDatasetByName(t *testing.T) {
	db := setupDB()
	defer db.Close()
	repo := database.NewRepository(db)
	cleanDB(repo)
	logger.Init()

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())

	svc := services.NewDatasetService(repo)
	user := testutils.CreateTestUser(t, repo, email)

	svc.CreateDataset(context.Background(), user.ID, "Finance Report", "")
	svc.CreateDataset(context.Background(), user.ID, "Engineering Data", "")

	results, err := svc.SearchDatasetByName(context.Background(), user.ID, "Finance", 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Name, "Finance")
}

func TestUploadDataset(t *testing.T) {
	db := setupDB()
	defer db.Close()
	repo := database.NewRepository(db)
	cleanDB(repo)
	logger.Init()

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())

	svc := services.NewDatasetService(repo)
	user := testutils.CreateTestUser(t, repo, email)

	content := []byte("id,value\n1,100\n2,200")
	reader := bytes.NewReader(content)

	dataset, err := svc.UploadDataset(context.Background(), user.ID, "upload.csv", reader)
	require.NoError(t, err)
	assert.Equal(t, "upload.csv", dataset.Name)
}
