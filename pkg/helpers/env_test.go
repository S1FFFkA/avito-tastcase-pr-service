package helpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		setValue     string
		defaultValue string
		want         string
	}{
		{
			name:         "environment variable is set",
			key:          "TEST_VAR",
			setValue:     "test_value",
			defaultValue: "default_value",
			want:         "test_value",
		},
		{
			name:         "environment variable is not set",
			key:          "TEST_VAR_NOT_SET",
			setValue:     "",
			defaultValue: "default_value",
			want:         "default_value",
		},
		{
			name:         "environment variable is empty string",
			key:          "TEST_VAR_EMPTY",
			setValue:     "",
			defaultValue: "default_value",
			want:         "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test
			defer os.Unsetenv(tt.key)

			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			}

			result := EnvOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, result)
		})
	}
}
