package helpers

import "testing"

func TestContainsReviewer(t *testing.T) {
	tests := []struct {
		name      string
		reviewers []string
		userID    string
		expected  bool
	}{
		{
			name:      "reviewer found",
			reviewers: []string{"user-1", "user-2", "user-3"},
			userID:    "user-2",
			expected:  true,
		},
		{
			name:      "reviewer not found",
			reviewers: []string{"user-1", "user-2", "user-3"},
			userID:    "user-4",
			expected:  false,
		},
		{
			name:      "empty reviewers list",
			reviewers: []string{},
			userID:    "user-1",
			expected:  false,
		},
		{
			name:      "single reviewer found",
			reviewers: []string{"user-1"},
			userID:    "user-1",
			expected:  true,
		},
		{
			name:      "single reviewer not found",
			reviewers: []string{"user-1"},
			userID:    "user-2",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsReviewer(tt.reviewers, tt.userID)
			if result != tt.expected {
				t.Errorf("ContainsReviewer() = %v, want %v", result, tt.expected)
			}
		})
	}
}
