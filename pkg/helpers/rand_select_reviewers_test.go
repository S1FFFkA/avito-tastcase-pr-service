package helpers

import (
	"AVITOSAMPISHU/internal/domain"
	"testing"
)

func TestRandSelectReviewers(t *testing.T) {
	tests := []struct {
		name      string
		members   []domain.TeamMember
		authorID  string
		maxCount  int
		validator func(t *testing.T, result []string)
	}{
		{
			name: "select all available candidates",
			members: []domain.TeamMember{
				{UserID: "user-1", IsActive: true},
				{UserID: "user-2", IsActive: true},
				{UserID: "user-3", IsActive: true},
			},
			authorID: "author-1",
			maxCount: 5,
			validator: func(t *testing.T, result []string) {
				if len(result) != 3 {
					t.Errorf("expected 3 reviewers, got %d", len(result))
				}
			},
		},
		{
			name: "select limited count",
			members: []domain.TeamMember{
				{UserID: "user-1", IsActive: true},
				{UserID: "user-2", IsActive: true},
				{UserID: "user-3", IsActive: true},
				{UserID: "user-4", IsActive: true},
			},
			authorID: "author-1",
			maxCount: 2,
			validator: func(t *testing.T, result []string) {
				if len(result) != 2 {
					t.Errorf("expected 2 reviewers, got %d", len(result))
				}
			},
		},
		{
			name: "exclude author",
			members: []domain.TeamMember{
				{UserID: "author-1", IsActive: true},
				{UserID: "user-1", IsActive: true},
				{UserID: "user-2", IsActive: true},
			},
			authorID: "author-1",
			maxCount: 5,
			validator: func(t *testing.T, result []string) {
				if len(result) != 2 {
					t.Errorf("expected 2 reviewers, got %d", len(result))
				}
				for _, reviewer := range result {
					if reviewer == "author-1" {
						t.Error("author should not be in reviewers list")
					}
				}
			},
		},
		{
			name: "exclude inactive members",
			members: []domain.TeamMember{
				{UserID: "user-1", IsActive: true},
				{UserID: "user-2", IsActive: false},
				{UserID: "user-3", IsActive: true},
			},
			authorID: "author-1",
			maxCount: 5,
			validator: func(t *testing.T, result []string) {
				if len(result) != 2 {
					t.Errorf("expected 2 reviewers, got %d", len(result))
				}
				for _, reviewer := range result {
					if reviewer == "user-2" {
						t.Error("inactive user should not be in reviewers list")
					}
				}
			},
		},
		{
			name: "zero max count",
			members: []domain.TeamMember{
				{UserID: "user-1", IsActive: true},
			},
			authorID: "author-1",
			maxCount: 0,
			validator: func(t *testing.T, result []string) {
				if len(result) != 0 {
					t.Errorf("expected 0 reviewers, got %d", len(result))
				}
			},
		},
		{
			name:     "empty members list",
			members:  []domain.TeamMember{},
			authorID: "author-1",
			maxCount: 5,
			validator: func(t *testing.T, result []string) {
				if len(result) != 0 {
					t.Errorf("expected 0 reviewers, got %d", len(result))
				}
			},
		},
		{
			name: "negative max count",
			members: []domain.TeamMember{
				{UserID: "user-1", IsActive: true},
			},
			authorID: "author-1",
			maxCount: -1,
			validator: func(t *testing.T, result []string) {
				if len(result) != 0 {
					t.Errorf("expected 0 reviewers, got %d", len(result))
				}
			},
		},
		{
			name: "all members inactive",
			members: []domain.TeamMember{
				{UserID: "user-1", IsActive: false},
				{UserID: "user-2", IsActive: false},
			},
			authorID: "author-1",
			maxCount: 5,
			validator: func(t *testing.T, result []string) {
				if len(result) != 0 {
					t.Errorf("expected 0 reviewers, got %d", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandSelectReviewers(tt.members, tt.authorID, tt.maxCount)
			tt.validator(t, result)
		})
	}
}
