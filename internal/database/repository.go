package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct {
	Queries *Queries
	DB      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Queries: New(db),
		DB:      db,
	}
}

func (r *Repository) Exec(query string, args ...interface{}) (sql.Result, error) {
	return r.DB.Exec(query, args...)
}

func (r *Repository) BatchInsertRecordValues(ctx context.Context, values []CreateRecordValueParams) error {
	if len(values) == 0 {
		return nil
	}

	var (
		queryBuilder strings.Builder
		args         []interface{}
	)

	queryBuilder.WriteString("INSERT INTO record_values (record_id, field_id, value) VALUES ")

	for i, v := range values {
		offset := i * 3
		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d)", offset+1, offset+2, offset+3))
		if i < len(values)-1 {
			queryBuilder.WriteString(", ")
		}
		args = append(args, v.RecordID, v.FieldID, v.Value)
	}

	query := queryBuilder.String()
	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}
