package services

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
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

func (s *DatasetService) UpdateDatasetRows(ctx context.Context, params database.UpdateDatasetRowsParams) error {
	err := s.Repo.Queries.UpdateDatasetRows(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update dataset rows: %w", err)
	}
	return nil
}

func (s *DatasetService) DeleteFieldFromDataset(c *gin.Context, datasetID uuid.UUID, fieldID uuid.UUID) error {
	// Call your database layer to delete the field from the dataset
	params := database.DeleteDatasetFieldParams{
		ID:        fieldID,
		DatasetID: datasetID,
	}

	// Perform the deletion
	err := s.Repo.Queries.DeleteDatasetField(context.Background(), params)
	return err
}

func (s *DatasetService) GetDatasetHeaders(ctx context.Context, datasetID, userID uuid.UUID) ([]string, error) {
	header, _, err := s.GetDatasetRows(ctx, datasetID, userID)
	if err != nil {
		return nil, err
	}
	if len(header) == 0 {
		return nil, fmt.Errorf("dataset is empty")
	}
	return header, nil
}

func (s *DatasetService) GetDatasetByID(ctx context.Context, id uuid.UUID) (database.Dataset, error) {
	dataset, err := s.Repo.Queries.GetDatasetByID(ctx, id)
	if err != nil {
		return database.Dataset{}, fmt.Errorf("error fetching dataset: %w", err)
	}
	return dataset, nil
}

func (s *DatasetService) GetDatasetRows(ctx context.Context, datasetID, userID uuid.UUID) ([]string, [][]string, error) {
	// 1. Load fields (columns)
	fields, err := s.Repo.Queries.GetFieldsByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get fields: %w", err)
	}

	// 2. Load records (rows)
	records, err := s.Repo.Queries.GetRecordsByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get records: %w", err)
	}

	if len(fields) == 0 || len(records) == 0 {
		return []string{}, [][]string{}, nil // empty dataset
	}

	// 3. Build header and field index map
	fieldIndexMap := make(map[uuid.UUID]int)
	header := make([]string, len(fields))
	for i, f := range fields {
		fieldIndexMap[f.ID] = i
		header[i] = f.Name
	}

	// 4. Load all record values
	values, err := s.Repo.Queries.GetRecordValuesByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get record values: %w", err)
	}

	// 5. Build value map per record with nil handling
	recordValueMap := make(map[uuid.UUID]map[int]*string)
	for _, val := range values {
		colIdx, ok := fieldIndexMap[val.FieldID]
		if !ok {
			continue
		}
		if _, exists := recordValueMap[val.RecordID]; !exists {
			recordValueMap[val.RecordID] = make(map[int]*string)
		}
		if val.Value.Valid {
			v := val.Value.String
			recordValueMap[val.RecordID][colIdx] = &v
		} else {
			recordValueMap[val.RecordID][colIdx] = nil
		}
	}

	// 6. Reconstruct all rows
	var rows [][]string
	for _, record := range records {
		row := make([]string, len(fields))
		colMap := recordValueMap[record.ID]

		for i := 0; i < len(fields); i++ {
			if colMap != nil {
				if val, ok := colMap[i]; ok && val != nil {
					row[i] = *val
				} else {
					row[i] = ""
				}
			} else {
				row[i] = ""
			}
		}
		rows = append(rows, row)
	}

	return header, rows, nil
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
		UserID:  userID,
		Column2: search,
		Limit:   limit,
		Offset:  offset,
	})
}

func (s *DatasetService) GetColumnsForDataset(ctx context.Context, datasetID uuid.UUID) ([]string, error) {
	return s.Repo.Queries.GetDatasetFieldsForDataset(ctx, datasetID)
}

func (s *DatasetService) UploadDataset(ctx context.Context, userID uuid.UUID, filename string, file io.Reader) (database.Dataset, error) {
	dataset, err := s.CreateDataset(ctx, userID, filename, "Uploaded dataset file")
	if err != nil {
		logger.Logger.Printf("Error uploading dataset: %v", err)
		return database.Dataset{}, err
	}

	csvReader := csv.NewReader(file)
	headers, err := csvReader.Read()
	if err != nil {
		return database.Dataset{}, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	const sampleLimit = 100

	// For type inference
	samples := make([][]string, len(headers))
	for i := range samples {
		samples[i] = []string{}
	}

	// Buffer all rows for later use
	var allRows [][]string
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return dataset, fmt.Errorf("error reading CSV row: %w", err)
		}

		allRows = append(allRows, row)

		// Collect samples up to sampleLimit
		if len(allRows) <= sampleLimit {
			for i, val := range row {
				if i < len(samples) {
					samples[i] = append(samples[i], val)
				}
			}
		}
	}

	// Infer types
	fieldTypes := make([]string, len(headers))
	for i := range headers {
		fieldTypes[i] = inferType(samples[i])
	}

	fieldIDs := make([]uuid.UUID, len(headers))
	for i, fieldName := range headers {
		fieldID := uuid.New()
		fieldIDs[i] = fieldID

		err := s.Repo.Queries.CreateDatasetField(ctx, database.CreateDatasetFieldParams{
			ID:          fieldID,
			DatasetID:   dataset.ID,
			Name:        fieldName,
			DataType:    fieldTypes[i],
			Description: sql.NullString{String: "", Valid: false},
			CreatedAt:   time.Now(),
		})

		if err != nil {
			return database.Dataset{}, fmt.Errorf("failed to insert dataset field: %w", err)
		}
	}

	for _, row := range allRows {
		recordID := uuid.New()
		err = s.Repo.Queries.CreateDatasetRecord(ctx, database.CreateDatasetRecordParams{
			ID:        recordID,
			DatasetID: dataset.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return dataset, fmt.Errorf("failed to insert dataset record: %w", err)
		}

		for i, val := range row {
			val = strings.TrimSpace(val)

			err := s.Repo.Queries.CreateRecordValue(ctx, database.CreateRecordValueParams{
				RecordID: recordID,
				FieldID:  fieldIDs[i],
				Value:    sql.NullString{String: val, Valid: val != ""},
			})
			if err != nil {
				return dataset, fmt.Errorf("failed to insert record value: %w", err)
			}
		}
	}

	return dataset, nil
}

func inferType(values []string) string {
	intCount := 0
	floatCount := 0
	boolCount := 0
	dateCount := 0
	total := len(values)

	for _, val := range values {
		val = strings.TrimSpace(val)
		if val == "" {
			total--
			continue
		}
		if _, err := strconv.Atoi(val); err == nil {
			intCount++
			continue
		}
		if _, err := strconv.ParseFloat(val, 64); err == nil {
			floatCount++
			continue
		}
		if _, err := strconv.ParseBool(val); err == nil {
			boolCount++
			continue
		}
		if isDate(val) {
			dateCount++
			continue
		}

	}

	if intCount == total {
		return "integer"
	}
	if floatCount+intCount >= total {
		return "float"
	}
	if boolCount >= total {
		return "boolean"
	}
	if dateCount >= total {
		return "datetime"
	}
	return "text"
}

func isDate(val string) bool {
	formats := []string{
		time.RFC3339, "2006-01-02", "01/02/2006", "02-Jan-2006", "Jan-02-2006", "02-Jan-06", "Jan-02-06",
	}
	for _, format := range formats {
		if _, err := time.Parse(format, val); err == nil {
			return true
		}
	}
	return false
}

func (s *DatasetService) GetNumericColumnValues(ctx context.Context, datasetID, userID uuid.UUID, column string) ([]float64, error) {
	headers, rows, err := s.GetDatasetRows(ctx, datasetID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset rows: %w", err)
	}

	// Find the index of the requested column
	colIdx := -1
	for i, h := range headers {
		if h == column {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return nil, fmt.Errorf("column '%s' not found in dataset", column)
	}

	var values []float64
	for i, row := range rows {
		if colIdx >= len(row) {
			continue // skip malformed rows
		}
		raw := row[colIdx]
		if raw == "" {
			continue
		}
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: cannot parse '%s' as float64", i, raw)
		}
		values = append(values, parsed)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no valid numeric values found in column '%s'", column)
	}

	return values, nil
}
