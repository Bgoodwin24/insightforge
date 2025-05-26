package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Dataset struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Public      bool      `json:"public"`
}

type DatasetHandler struct {
	Service *services.DatasetService
}

func NewDatasetHandler(service *services.DatasetService) *DatasetHandler {
	return &DatasetHandler{
		Service: service,
	}
}

func (h *DatasetHandler) CheckDatasetOwnership(c *gin.Context, datasetID uuid.UUID) (*database.Dataset, bool) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		return nil, false
	}

	dataset, err := h.Service.GetDatasetByIDForUser(c, userID, datasetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
		return nil, false
	}

	// Check if dataset belongs to the user
	if dataset.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		return nil, false
	}

	return &dataset, true
}

func (h *DatasetHandler) CreateDataset(c *gin.Context) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dataset name is required"})
		return
	}

	userID, ok := GetUserIDFromContext(c)
	if !ok {
		return
	}

	dataset, err := h.Service.CreateDataset(c, userID, input.Name, input.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dataset"})
		return
	}
	dataset.CreatedAt = time.Now()
	dataset.UpdatedAt = time.Now()

	c.JSON(http.StatusCreated, dataset)
}

func (h *DatasetHandler) ListDatasets(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	if limit > 1000 {
		limit = 1000
	}
	offset, _ := strconv.Atoi(offsetStr)
	if offset < 0 {
		offset = 0
	}

	datasets, err := h.Service.ListDatasetsForUser(c, userID, int32(limit), int32(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list datasets"})
		return
	}

	c.JSON(http.StatusOK, datasets)
}

func (h *DatasetHandler) SearchDataSets(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		return
	}

	search := c.DefaultQuery("search", "")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	if limit > 1000 {
		limit = 1000
	}
	offset, _ := strconv.Atoi(offsetStr)
	if offset < 0 {
		offset = 0
	}

	datasets, err := h.Service.SearchDatasetByName(c, userID, search, int32(limit), int32(offset))
	if err != nil {
		log.Printf("Error in SearchDatasetByName: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, datasets)
}

func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user session"})
		return uuid.Nil, false
	}

	return userID, true
}

type DatasetResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Columns []string  `json:"columns"`
}

func (h *DatasetHandler) UploadDataset(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file upload failed"})
		return
	}
	defer file.Close()

	limitSizeStr := os.Getenv("MAX_UPLOAD_SIZE")
	if limitSizeStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MAX_UPLOAD_SIZE is not set in .env"})
		return
	}
	limitSize, err := strconv.ParseInt(limitSizeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid MAX_UPLOAD_SIZE setting"})
		return
	}

	if header.Size > limitSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	dataset, err := h.Service.UploadDataset(c, userID, header.Filename, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to upload dataset: %v", err)})
		return
	}

	columns, err := h.Service.GetColumnsForDataset(c, dataset.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get dataset columns"})
		return
	}

	c.JSON(http.StatusCreated, DatasetResponse{
		ID:      dataset.ID,
		Name:    dataset.Name,
		Columns: columns,
	})
}

func (h *DatasetHandler) GetDatasetByID(c *gin.Context) {
	idStr := c.Param("id")
	datasetID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dataset ID"})
		return
	}

	dataset, authorized := h.CheckDatasetOwnership(c, datasetID)
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized to access this dataset"})
		return
	}

	columns, err := h.Service.GetColumnsForDataset(c.Request.Context(), datasetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching dataset columns"})
		return
	}

	userID, ok := GetUserIDFromContext(c)
	if !ok {
		return
	}

	_, rows, err := h.Service.GetDatasetRows(c.Request.Context(), datasetID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         dataset.ID,
		"name":       dataset.Name,
		"created_at": dataset.CreatedAt,
		"columns":    columns,
		"rows":       rows,
	})
}

func (h *DatasetHandler) DeleteDatasetsByID(c *gin.Context) {
	idStr := c.Param("id")
	datasetID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dataset ID"})
		return
	}

	dataset, authorized := h.CheckDatasetOwnership(c, datasetID)
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized to delete this dataset"})
		return
	}

	err = h.Service.DeleteDataset(c, datasetID, dataset.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete dataset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dataset deleted"})
}

func (h *DatasetHandler) UpdateDataset(c *gin.Context) {
	idStr := c.Param("id")
	datasetID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dataset ID"})
		return
	}

	dataset, err := h.Service.Repo.Queries.GetDatasetByID(c, datasetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dataset name is required"})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authorized"})
		return
	}

	userIDParsed, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user_id"})
		return
	}

	if dataset.UserID != userIDParsed {
		log.Printf("User is not authorized to update dataset %s", datasetID)
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	updated, err := h.Service.UpdateDataset(c, datasetID, input.Name, input.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	resp := Dataset{
		ID:          updated.ID,
		Name:        updated.Name,
		Description: nullStringToStr(updated.Description),
		CreatedAt:   updated.CreatedAt,
		UpdatedAt:   updated.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

func nullStringToStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
