package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		setEnv       bool
		defaultValue string
		expected     string
	}{
		{
			name:         "success get value",
			envKey:       "test key",
			envValue:     "test value",
			setEnv:       true,
			defaultValue: "default",
			expected:     "test value",
		},
		{
			name:         "success get default",
			envKey:       "nonexistent value",
			setEnv:       false,
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got := EnvWithDefault(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEnvWithDefaultBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		setEnv       bool
		defaultValue bool
		expected     bool
	}{
		{
			name:         "success get value",
			envKey:       "test key",
			envValue:     "true",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "success get default",
			envKey:       "nonexistent value",
			setEnv:       false,
			defaultValue: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got := EnvWithDefaultBool(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEnvWithDefaultFloat(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		setEnv       bool
		defaultValue float64
		expected     float64
	}{
		{
			name:         "success get value",
			envKey:       "test float key",
			envValue:     "0.75",
			setEnv:       true,
			defaultValue: 0.5,
			expected:     0.75,
		},
		{
			name:         "success get default",
			envKey:       "nonexistent float value",
			setEnv:       false,
			defaultValue: 1.5,
			expected:     1.5,
		},
		{
			name:         "parse negative float",
			envKey:       "negative float key",
			envValue:     "-0.5",
			setEnv:       true,
			defaultValue: 0.0,
			expected:     -0.5,
		},
		{
			name:         "parse integer as float",
			envKey:       "int as float key",
			envValue:     "5",
			setEnv:       true,
			defaultValue: 0.0,
			expected:     5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got := EnvWithDefaultFloat(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, got)
		})
	}
}
