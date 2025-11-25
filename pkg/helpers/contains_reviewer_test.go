package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsReviewer(t *testing.T) {
	tests := []struct {
		name      string
		reviewers []string
		userID    string
		want      bool
	}{
		{
			name:      "reviewer is in list",
			reviewers: []string{"user1", "user2", "user3"},
			userID:    "user2",
			want:      true,
		},
		{
			name:      "reviewer is not in list",
			reviewers: []string{"user1", "user2", "user3"},
			userID:    "user4",
			want:      false,
		},
		{
			name:      "empty list",
			reviewers: []string{},
			userID:    "user1",
			want:      false,
		},
		{
			name:      "single element list - found",
			reviewers: []string{"user1"},
			userID:    "user1",
			want:      true,
		},
		{
			name:      "single element list - not found",
			reviewers: []string{"user1"},
			userID:    "user2",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsReviewer(tt.reviewers, tt.userID)
			assert.Equal(t, tt.want, result)
		})
	}
}
