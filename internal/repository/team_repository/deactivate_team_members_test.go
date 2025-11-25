package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logger.InitLogger()
}

func TestTeamStorage_DeactivateTeamMembers(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		userIDs       []string
		reassignments []domain.ReviewerReassignment
		setup         func(mock sqlmock.Sqlmock)
		want          []string
		wantErr       error
	}{
		{
			name:          "successful deactivation with specific users",
			teamName:      "team1",
			userIDs:       []string{"user1", "user2"},
			reassignments: []domain.ReviewerReassignment{},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).
					AddRow("user1").
					AddRow("user2")
				mock.ExpectQuery(`UPDATE users u`).
					WithArgs("team1", pq.Array([]string{"user1", "user2"})).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
			want:    []string{"user1", "user2"},
			wantErr: nil,
		},
		{
			name:     "successful deactivation with reassignments",
			teamName: "team1",
			userIDs:  []string{"user1"},
			reassignments: []domain.ReviewerReassignment{
				{PrID: "pr1", OldReviewerID: "user1", NewReviewerID: "user3"},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow("user1")
				mock.ExpectQuery(`UPDATE users u`).
					WithArgs("team1", pq.Array([]string{"user1"})).
					WillReturnRows(rows)
				mock.ExpectExec(`DELETE FROM reviewers`).
					WithArgs("pr1", "user1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO reviewers`).
					WithArgs("pr1", "user3").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			want:    []string{"user1"},
			wantErr: nil,
		},
		{
			name:          "database error on update",
			teamName:      "team1",
			userIDs:       []string{"user1"},
			reassignments: []domain.ReviewerReassignment{},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE users u`).
					WithArgs("team1", pq.Array([]string{"user1"})).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			want:    nil,
			wantErr: sql.ErrConnDone,
		},
		{
			name:     "foreign key error on reassignment",
			teamName: "team1",
			userIDs:  []string{"user1"},
			reassignments: []domain.ReviewerReassignment{
				{PrID: "pr1", OldReviewerID: "user1", NewReviewerID: "non-existent"},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow("user1")
				mock.ExpectQuery(`UPDATE users u`).
					WithArgs("team1", pq.Array([]string{"user1"})).
					WillReturnRows(rows)
				mock.ExpectExec(`DELETE FROM reviewers`).
					WithArgs("pr1", "user1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				pqErr := &pq.Error{Code: "23503"}
				mock.ExpectExec(`INSERT INTO reviewers`).
					WithArgs("pr1", "non-existent").
					WillReturnError(pqErr)
				mock.ExpectRollback()
			},
			want:    nil,
			wantErr: &pq.Error{Code: "23503"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setup(mock)

			repo := NewTeamStorage(db)
			got, err := repo.DeactivateTeamMembers(context.Background(), tt.teamName, tt.userIDs, tt.reassignments)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if pqErr, ok := tt.wantErr.(*pq.Error); ok {
					if gotPqErr, ok := err.(*pq.Error); ok {
						assert.Equal(t, pqErr.Code, gotPqErr.Code)
					} else {
						assert.Contains(t, err.Error(), "foreign key")
					}
				} else {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
