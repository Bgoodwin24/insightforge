package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/google/uuid"
)

type DatasetService struct {
	Repo *database.Repository
}

func NewDatasetService(repo *database.Repository) *DatasetService {
	return &DatasetService{
		Repo: repo,
	}
}

func (s *DatasetService) CreateDataset(ctx context.Context, userID uuid.UUID, name, description string) (database.Dataset, error) {
	now := time.Now()

	return s.Repo.Queries.CreateDataset(ctx, database.CreateDatasetParams{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *DatasetService) GetDatasetByIDForUser(ctx context.Context, userID, datasetID uuid.UUID) (database.Dataset, error) {
	dataset, err := s.Repo.Queries.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return database.Dataset{}, err
	}
	if dataset.UserID != userID {
		return database.Dataset{}, fmt.Errorf("unauthorized access")
	}
	return dataset, nil
}

func (s *DatasetService) ListDatasetsForUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]database.Dataset, error) {
	return s.Repo.Queries.ListDatasetsForUser(ctx, database.ListDatasetsForUserParams{
		Limit:  limit,
		Offset: offset,
		ID:     userID,
	})
}

func (s *DatasetService) UpdateDataset(ctx context.Context, id uuid.UUID, name, description string) (database.Dataset, error) {
	now := time.Now()
	return s.Repo.Queries.UpdateDataset(ctx, database.UpdateDatasetParams{
		ID:          id,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		UpdatedAt:   now,
	})
}

func (s *DatasetService) DeleteDataset(ctx context.Context, id, userID uuid.UUID) error {
	return s.Repo.Queries.DeleteDataset(ctx, database.DeleteDatasetParams{
		ID:     id,
		UserID: userID,
	})
}

func (s *DatasetService) SearchDatasetByName(ctx context.Context, userID uuid.UUID, search string, limit, offset int32) ([]database.Dataset, error) {
	return s.Repo.Queries.SearchDatasetByName(ctx, database.SearchDatasetByNameParams{
		UserID: userID,
		Search: sql.NullString{String: search, Valid: search != ""},
		Limit:  limit,
		Offset: offset,
	})
}

func (s *DatasetService) UploadDataset(ctx context.Context, userID uuid.UUID, filename string, file io.Reader) (database.Dataset, error) {
	// Storage setup
	//use sql db

	dataset, err := s.CreateDataset(ctx, userID, filename, "Uploaded dataset file")
	if err != nil {
		logger.Logger.Printf("Error uploading dataset: %v", err)
		return database.Dataset{}, err
	}
	return dataset, nil
}
