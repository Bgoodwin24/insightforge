package database

type Repository struct {
	Queries *Queries
}

func NewRepository(db DBTX) *Repository {
	return &Repository{
		Queries: New(db),
	}
}
