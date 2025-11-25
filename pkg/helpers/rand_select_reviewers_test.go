package helpers

import (
	"AVITOSAMPISHU/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandSelectReviewers(t *testing.T) {
	tests := []struct {
		name           string
		members        []domain.TeamMember
		authorID       string
		maxCount       int
		wantCount      int
		wantNoAuthor   bool
		wantOnlyActive bool
	}{
		{
			name: "normal case - select 2 from 5",
			members: []domain.TeamMember{
				{UserID: "user1", IsActive: true},
				{UserID: "user2", IsActive: true},
				{UserID: "user3", IsActive: true},
				{UserID: "user4", IsActive: true},
				{UserID: "user5", IsActive: true},
			},
			authorID:       "user1",
			maxCount:       2,
			wantCount:      2,
			wantNoAuthor:   true,
			wantOnlyActive: true,
		},
		{
			name: "exclude inactive users",
			members: []domain.TeamMember{
				{UserID: "user1", IsActive: true},
				{UserID: "user2", IsActive: false},
				{UserID: "user3", IsActive: true},
				{UserID: "user4", IsActive: false},
			},
			authorID:       "user1",
			maxCount:       2,
			wantCount:      1, // Only user3 available (user1 is author, user2 and user4 are inactive)
			wantNoAuthor:   true,
			wantOnlyActive: true,
		},
		{
			name: "maxCount greater than available candidates",
			members: []domain.TeamMember{
				{UserID: "user1", IsActive: true},
				{UserID: "user2", IsActive: true},
			},
			authorID:       "user1",
			maxCount:       5,
			wantCount:      1, // Only user2 available
			wantNoAuthor:   true,
			wantOnlyActive: true,
		},
		{
			name: "maxCount is zero",
			members: []domain.TeamMember{
				{UserID: "user1", IsActive: true},
				{UserID: "user2", IsActive: true},
			},
			authorID:       "user1",
			maxCount:       0,
			wantCount:      0,
			wantNoAuthor:   true,
			wantOnlyActive: true,
		},
		{
			name: "all users inactive",
			members: []domain.TeamMember{
				{UserID: "user1", IsActive: false},
				{UserID: "user2", IsActive: false},
			},
			authorID:       "user1",
			maxCount:       2,
			wantCount:      0,
			wantNoAuthor:   true,
			wantOnlyActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandSelectReviewers(tt.members, tt.authorID, tt.maxCount)

			require.Len(t, result, tt.wantCount, "result length should match expected count")

			if tt.wantNoAuthor {
				assert.NotContains(t, result, tt.authorID, "result should not contain author")
			}

			if tt.wantOnlyActive {
				memberMap := make(map[string]bool)
				for _, m := range tt.members {
					memberMap[m.UserID] = m.IsActive
				}
				for _, reviewerID := range result {
					assert.True(t, memberMap[reviewerID], "selected reviewer should be active")
				}
			}
		})
	}
}
