package repository

import (
	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/pkg/logger"
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logger.InitLogger()
}

func TestUserRepository_GetUserByID(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		setup   func(mock sqlmock.Sqlmock)
		want    *domain.User
		wantErr error
	}{
		{
			name:   "successful get",
			userID: "user1",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username", "team_name", "is_active"}).
					AddRow("User1", "Team1", true)
				mock.ExpectQuery(`SELECT u.username, t.team_name, u.is_active`).
					WithArgs("user1").
					WillReturnRows(rows)
			},
			want: &domain.User{
				UserID:   "user1",
				Username: "User1",
				TeamName: "Team1",
				IsActive: true,
			},
			wantErr: nil,
		},
		{
			name:   "user not found",
			userID: "user999",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT u.username, t.team_name, u.is_active`).
					WithArgs("user999").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: domain.ErrNotFound,
		},
		{
			name:   "database error",
			userID: "user1",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT u.username, t.team_name, u.is_active`).
					WithArgs("user1").
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.setup(mock)

			repo := NewUserRepository(db)
			got, err := repo.GetUserByID(context.Background(), tt.userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
