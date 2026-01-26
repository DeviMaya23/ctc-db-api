package helpers

import (
	"fmt"
	"lizobly/ctc-db-api/pkg/constants"
	"net/http"

	"github.com/labstack/echo/v4"
)

// SetCacheHeaders sets Cache-Control, ETag, and Last-Modified headers for a resource.
// Returns true if the client's cached version is still valid (304 Not Modified should be returned).
func SetCacheHeaders(ctx echo.Context, etag string, lastModified string, maxAge int) bool {
	// Set cache headers
	ctx.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	ctx.Response().Header().Set("ETag", etag)
	ctx.Response().Header().Set("Last-Modified", lastModified)

	// Check if client has valid cached version
	if ctx.Request().Header.Get("If-None-Match") == etag {
		return true
	}

	return false
}

// SetListCacheHeaders sets Cache-Control header for list endpoints.
func SetListCacheHeaders(ctx echo.Context) {
	ctx.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", constants.CacheMaxAgeList))
}

// CheckETagMatch checks if the client's If-Match header matches the current ETag.
// Returns true if they match or if no If-Match header is provided.
// Returns false if there's a mismatch, indicating the resource was modified by another request.
func CheckETagMatch(ctx echo.Context, currentETag string) bool {
	clientETag := ctx.Request().Header.Get("If-Match")
	if clientETag == "" {
		// No If-Match header provided, allow the operation
		return true
	}

	return clientETag == currentETag
}

// RespondNotModified sends a 304 Not Modified response.
func RespondNotModified(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotModified)
}

// RespondPreconditionFailed sends a 412 Precondition Failed response with a message.
func RespondPreconditionFailed(ctx echo.Context) error {
	return ctx.JSON(http.StatusPreconditionFailed, map[string]string{
		"error": "Resource has been modified by another request. Please refresh and try again.",
	})
}
