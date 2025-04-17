package database

import "database/sql"

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

// Add more methods that might combine multiple queries
