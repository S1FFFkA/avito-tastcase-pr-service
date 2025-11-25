package repository

import (
	"database/sql"
)

type PullRequestStorage struct {
	db *sql.DB
}

func NewPullRequestStorage(db *sql.DB) *PullRequestStorage {
	return &PullRequestStorage{
		db: db,
	}
}
