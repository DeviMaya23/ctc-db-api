package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSetCacheHeaders(t *testing.T) {
	tests := []struct {
		name             string
		etag             string
		lastModified     string
		maxAge           int
		ifNoneMatchValue string
		expectedResult   bool
		expectedCC       string
	}{
		{
			name:             "sets cache headers correctly",
			etag:             `"abc123"`,
			lastModified:     "Mon, 27 Jan 2026 10:00:00 GMT",
			maxAge:           3600,
			ifNoneMatchValue: "",
			expectedResult:   false,
			expectedCC:       "public, max-age=3600",
		},
		{
			name:             "returns true when client has valid cached version",
			etag:             `"abc123"`,
			lastModified:     "Mon, 27 Jan 2026 10:00:00 GMT",
			maxAge:           3600,
			ifNoneMatchValue: `"abc123"`,
			expectedResult:   true,
			expectedCC:       "public, max-age=3600",
		},
		{
			name:             "returns false when ETag does not match",
			etag:             `"new456"`,
			lastModified:     "Mon, 27 Jan 2026 10:00:00 GMT",
			maxAge:           3600,
			ifNoneMatchValue: `"old123"`,
			expectedResult:   false,
			expectedCC:       "public, max-age=3600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.ifNoneMatchValue != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatchValue)
			}
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			result := SetCacheHeaders(ctx, tt.etag, tt.lastModified, tt.maxAge)

			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedCC, rec.Header().Get("Cache-Control"))
			assert.Equal(t, tt.etag, rec.Header().Get("ETag"))
			assert.Equal(t, tt.lastModified, rec.Header().Get("Last-Modified"))
		})
	}
}

func TestSetListCacheHeaders(t *testing.T) {
	t.Run("sets list cache headers", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		SetListCacheHeaders(ctx)

		assert.Contains(t, rec.Header().Get("Cache-Control"), "public, max-age=")
	})
}

func TestCheckETagMatch(t *testing.T) {
	tests := []struct {
		name           string
		ifMatchHeader  string
		currentETag    string
		expectedResult bool
	}{
		{
			name:           "returns true when If-Match matches ETag",
			ifMatchHeader:  `"abc123"`,
			currentETag:    `"abc123"`,
			expectedResult: true,
		},
		{
			name:           "returns false when If-Match does not match ETag",
			ifMatchHeader:  `"old123"`,
			currentETag:    `"new456"`,
			expectedResult: false,
		},
		{
			name:           "returns true when no If-Match header provided",
			ifMatchHeader:  "",
			currentETag:    `"abc123"`,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/", nil)
			if tt.ifMatchHeader != "" {
				req.Header.Set("If-Match", tt.ifMatchHeader)
			}
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			result := CheckETagMatch(ctx, tt.currentETag)

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestRespondNotModified(t *testing.T) {
	t.Run("returns 304 Not Modified", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		err := RespondNotModified(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotModified, rec.Code)
	})
}

func TestRespondPreconditionFailed(t *testing.T) {
	t.Run("returns 412 Precondition Failed with message", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		err := RespondPreconditionFailed(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)
		assert.Contains(t, rec.Body.String(), "Resource has been modified")
	})
}
