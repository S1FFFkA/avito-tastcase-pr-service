package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logger.InitLogger()
}

func TestTeamStorage_CreateTeamWithMembers(t *testing.T) {
	tests := []struct {
		name      string
		teamName  string
		members   []domain.TeamMember
		setup     func(mock sqlmock.Sqlmock)
		wantErr   error
		checkTeam bool
	}{
		{
			name:     "successful creation",
			teamName: "team1",
			members: []domain.TeamMember{
				{UserID: "user1", Username: "User1", IsActive: true},
				{UserID: "user2", Username: "User2", IsActive: true},
			},
			setup: func(mock sqlmock.Sqlmock) {
				teamID := uuid.New()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO teams`).
					WithArgs("team1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(teamID))
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs("user1", "User1", teamID, true).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs("user2", "User2", teamID, true).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr:   nil,
			checkTeam: true,
		},
		{
			name:     "team already exists",
			teamName: "existing-team",
			members: []domain.TeamMember{
				{UserID: "user1", Username: "User1", IsActive: true},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				pqErr := &pq.Error{Code: "23505"}
				mock.ExpectQuery(`INSERT INTO teams`).
					WithArgs("existing-team").
					WillReturnError(pqErr)
				mock.ExpectRollback()
			},
			wantErr:   domain.ErrTeamExists,
			checkTeam: false,
		},
		{
			name:     "database error on user insert",
			teamName: "team1",
			members: []domain.TeamMember{
				{UserID: "user1", Username: "User1", IsActive: true},
			},
			setup: func(mock sqlmock.Sqlmock) {
				teamID := uuid.New()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO teams`).
					WithArgs("team1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(teamID))
				mock.ExpectExec(`INSERT INTO users`).
					WithArgs("user1", "User1", teamID, true).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr:   sql.ErrConnDone,
			checkTeam: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setup(mock)

			repo := NewTeamStorage(db)
			got, err := repo.CreateTeamWithMembers(context.Background(), tt.teamName, tt.members)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, uuid.Nil, got)
			} else {
				assert.NoError(t, err)
				if tt.checkTeam {
					assert.NotEqual(t, uuid.Nil, got)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
