package repository

import (
	"database/sql"
)

type PrReviewersStorage struct {
	db *sql.DB
}

func NewPrReviewersStorage(db *sql.DB) *PrReviewersStorage {
	return &PrReviewersStorage{
		db: db,
	}
}
