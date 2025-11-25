package repository

import (
	"database/sql"
)

type TeamStorage struct {
	db *sql.DB
}

func NewTeamStorage(db *sql.DB) *TeamStorage {
	return &TeamStorage{
		db: db,
	}
}
