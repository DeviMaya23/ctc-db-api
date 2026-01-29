package domain

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCommonModel_ETag tests ETag generation with various scenarios
func TestCommonModel_ETag(t *testing.T) {
	tests := []struct {
		name        string
		model       CommonModel
		validations func(t *testing.T, eTag string)
	}{
		{
			name: "generates quoted timestamp",
			model: CommonModel{
				ID:        1,
				CreatedBy: "user",
				UpdatedAt: time.Now(),
			},
			validations: func(t *testing.T, eTag string) {
				assert.NotEmpty(t, eTag)
				assert.True(t, eTag[0] == '"')
				assert.True(t, eTag[len(eTag)-1] == '"')
			},
		},
		{
			name: "properly formatted",
			model: CommonModel{
				ID:        42,
				UpdatedAt: time.Date(2025, 6, 15, 10, 30, 45, 0, time.UTC),
			},
			validations: func(t *testing.T, eTag string) {
				assert.Equal(t, '"', rune(eTag[0]))
				assert.Equal(t, '"', rune(eTag[len(eTag)-1]))
			},
		},
		{
			name:  "zero value model",
			model: CommonModel{},
			validations: func(t *testing.T, eTag string) {
				assert.NotEmpty(t, eTag)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eTag := tt.model.ETag()
			tt.validations(t, eTag)
		})
	}
}

// TestCommonModel_ETag_WithDifferentTimestamps tests ETag changes with UpdatedAt
func TestCommonModel_ETag_WithDifferentTimestamps(t *testing.T) {
	time1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)

	model1 := CommonModel{ID: 1, UpdatedAt: time1}
	model2 := CommonModel{ID: 1, UpdatedAt: time2}

	eTag1 := model1.ETag()
	eTag2 := model2.ETag()

	// Different timestamps should produce different ETags
	assert.NotEqual(t, eTag1, eTag2)
}

// TestCommonModel_ETag_Consistency tests ETag is consistent for same model
func TestCommonModel_ETag_Consistency(t *testing.T) {
	now := time.Now()
	model := CommonModel{
		ID:        1,
		UpdatedAt: now,
	}

	eTag1 := model.ETag()
	eTag2 := model.ETag()

	// Same model should produce same ETag
	assert.Equal(t, eTag1, eTag2)
}

// TestCommonModel_LastModified tests LastModified HTTP date format
func TestCommonModel_LastModified(t *testing.T) {
	tests := []struct {
		name        string
		model       CommonModel
		validations func(t *testing.T, lastModified string)
	}{
		{
			name: "returns RFC7231 format",
			model: CommonModel{
				ID:        1,
				UpdatedAt: time.Date(2025, 1, 15, 10, 30, 45, 0, time.UTC),
			},
			validations: func(t *testing.T, lastModified string) {
				_, err := time.Parse(http.TimeFormat, lastModified)
				assert.NoError(t, err)
			},
		},
		{
			name: "is in UTC timezone",
			model: func() CommonModel {
				loc, _ := time.LoadLocation("America/New_York")
				return CommonModel{
					ID:        1,
					UpdatedAt: time.Date(2025, 1, 15, 10, 30, 45, 0, loc),
				}
			}(),
			validations: func(t *testing.T, lastModified string) {
				parsed, err := time.Parse(http.TimeFormat, lastModified)
				assert.NoError(t, err)
				assert.Equal(t, "UTC", parsed.Location().String())
			},
		},
		{
			name:  "zero value model",
			model: CommonModel{},
			validations: func(t *testing.T, lastModified string) {
				assert.NotEmpty(t, lastModified)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastModified := tt.model.LastModified()
			tt.validations(t, lastModified)
		})
	}
}

// TestCommonModel_LastModified_Consistency tests LastModified is consistent
func TestCommonModel_LastModified_Consistency(t *testing.T) {
	now := time.Now()
	model := CommonModel{
		ID:        1,
		UpdatedAt: now,
	}

	lastMod1 := model.LastModified()
	lastMod2 := model.LastModified()

	// Same model should produce same LastModified value
	assert.Equal(t, lastMod1, lastMod2)
}

// TestCommonModel_LastModified_WithMultipleTimes tests consistency across different times
func TestCommonModel_LastModified_WithMultipleTimes(t *testing.T) {
	times := []time.Time{
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC),
		time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	for _, testTime := range times {
		model := CommonModel{
			ID:        1,
			UpdatedAt: testTime,
		}

		lastModified := model.LastModified()

		// All should be valid HTTP dates
		_, err := time.Parse(http.TimeFormat, lastModified)
		assert.NoError(t, err, "time %v should produce valid HTTP date format", testTime)
	}
}
