package helpers

import (
	"database/sql/driver"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockDB(t *testing.T) {
	db, mock, err := NewMockDB()

	require.NoError(t, err, "should create mock DB without error")
	assert.NotNil(t, db, "gorm DB should not be nil")
	assert.NotNil(t, mock, "sqlmock should not be nil")

	// Verify we can set expectations
	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows([]string{"id"}))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()
}

func TestAnyTime_Match(t *testing.T) {
	matcher := AnyTime{}

	tests := []struct {
		name     string
		value    driver.Value
		expected bool
	}{
		{"valid time.Time", time.Now(), true},
		{"zero time", time.Time{}, true},
		{"string value", "2024-01-01", false},
		{"int value", 123, false},
		{"nil value", nil, false},
		{"float value", 3.14, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetHTTPTestRecorder(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		requestBody interface{}
		queryParams url.Values
		pathParams  map[string]string
	}{
		{
			name:        "simple GET request",
			method:      http.MethodGet,
			url:         "/api/v1/users",
			requestBody: nil,
			queryParams: nil,
			pathParams:  nil,
		},
		{
			name:   "POST with JSON body",
			method: http.MethodPost,
			url:    "/api/v1/users",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
			queryParams: nil,
			pathParams:  nil,
		},
		{
			name:        "GET with query params",
			method:      http.MethodGet,
			url:         "/api/v1/users",
			requestBody: nil,
			queryParams: url.Values{
				"page":     []string{"1"},
				"pageSize": []string{"10"},
			},
			pathParams: nil,
		},
		{
			name:        "GET with path params",
			method:      http.MethodGet,
			url:         "/api/v1/users/:id",
			requestBody: nil,
			queryParams: nil,
			pathParams: map[string]string{
				"id": "123",
			},
		},
		{
			name:   "PUT with all params",
			method: http.MethodPut,
			url:    "/api/v1/users/:id",
			requestBody: map[string]interface{}{
				"username": "updated",
			},
			queryParams: url.Values{
				"force": []string{"true"},
			},
			pathParams: map[string]string{
				"id": "456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, ctx := GetHTTPTestRecorder(t, tt.method, tt.url, tt.requestBody, tt.queryParams, tt.pathParams)

			assert.NotNil(t, rec, "recorder should not be nil")
			assert.NotNil(t, ctx, "context should not be nil")

			// Verify method
			assert.Equal(t, tt.method, ctx.Request().Method)

			// Verify query params
			if tt.queryParams != nil {
				for key, values := range tt.queryParams {
					for _, value := range values {
						assert.Equal(t, value, ctx.QueryParam(key))
					}
				}
			}

			// Verify path params
			if tt.pathParams != nil {
				for key, value := range tt.pathParams {
					assert.Equal(t, value, ctx.Param(key))
				}
			}

			// Verify content type if request body exists
			if tt.requestBody != nil {
				assert.Equal(t, echo.MIMEApplicationJSON, ctx.Request().Header.Get(echo.HeaderContentType))
			}

			// Verify validator is set
			assert.NotNil(t, ctx.Get("validator"))
		})
	}
}

func TestGetHTTPTestRecorder_MultipleQueryValues(t *testing.T) {
	queryParams := url.Values{
		"tags": []string{"tag1", "tag2", "tag3"},
	}

	rec, ctx := GetHTTPTestRecorder(t, http.MethodGet, "/api/v1/items", nil, queryParams, nil)

	assert.NotNil(t, rec)
	assert.NotNil(t, ctx)

	// Verify all query values are present
	assert.Contains(t, ctx.Request().URL.RawQuery, "tags=tag1")
	assert.Contains(t, ctx.Request().URL.RawQuery, "tags=tag2")
	assert.Contains(t, ctx.Request().URL.RawQuery, "tags=tag3")
}
