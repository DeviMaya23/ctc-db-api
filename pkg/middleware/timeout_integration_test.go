package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pgGormDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTimeoutMiddleware_Integration_DatabaseQueryCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// ctx := context.Background()

	// Setup test database
	connStr := helpers.GetTestDB(t)
	dbConn, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	defer dbConn.Close()

	db, err := gorm.Open(pgGormDriver.New(pgGormDriver.Config{
		Conn: dbConn,
	}), &gorm.Config{
		TranslateError: true,
	})
	require.NoError(t, err)

	logger, _ := logging.NewDevelopmentLogger()

	t.Run("long query gets cancelled by context timeout", func(t *testing.T) {
		e := echo.New()

		// Handler that executes a long-running query
		slowQueryHandler := func(c echo.Context) error {
			ctx := c.Request().Context()

			// This query uses pg_sleep to simulate a long-running operation
			// Context cancellation should interrupt this
			var result int
			err := db.WithContext(ctx).Raw("SELECT pg_sleep(5)").Scan(&result).Error

			if err != nil {
				// Check if error is due to context cancellation
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			}

			return c.String(http.StatusOK, "query completed")
		}

		// Use very short timeout to force cancellation
		middleware := TimeoutMiddleware(100*time.Millisecond, logger)
		handler := middleware(slowQueryHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		echoCtx := e.NewContext(req, rec)

		err := handler(echoCtx)
		require.NoError(t, err)

		// Should return timeout response
		assert.Equal(t, http.StatusRequestTimeout, rec.Code)

		var response controller.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "request timeout", response.Message)
	})

	t.Run("transaction rolls back on timeout", func(t *testing.T) {
		e := echo.New()

		// Create a test table for transaction testing
		err := db.Exec(`
			CREATE TABLE IF NOT EXISTS test_timeout_table (
				id SERIAL PRIMARY KEY,
				value TEXT
			)
		`).Error
		require.NoError(t, err)
		defer func() {
			db.Exec("DROP TABLE IF EXISTS test_timeout_table")
		}()

		// Clear any existing data
		err = db.Exec("DELETE FROM test_timeout_table").Error
		require.NoError(t, err)

		// Handler that starts transaction and times out
		transactionHandler := func(c echo.Context) error {
			ctx := c.Request().Context()

			return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				// Insert a record
				err := tx.Exec("INSERT INTO test_timeout_table (value) VALUES (?)", "test_value").Error
				if err != nil {
					return err
				}

				// Simulate long operation that will timeout
				var result int
				err = tx.Raw("SELECT pg_sleep(5)").Scan(&result).Error
				if err != nil {
					return err
				}

				return nil
			})
		}

		middleware := TimeoutMiddleware(100*time.Millisecond, logger)
		handler := middleware(transactionHandler)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		echoCtx := e.NewContext(req, rec)

		err = handler(echoCtx)
		require.NoError(t, err)

		// Should return timeout response
		assert.Equal(t, http.StatusRequestTimeout, rec.Code)

		// Give time for transaction to complete/rollback
		time.Sleep(200 * time.Millisecond)

		// Verify no data was inserted (transaction rolled back)
		var count int64
		err = db.Raw("SELECT COUNT(*) FROM test_timeout_table").Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "transaction should have been rolled back")
	})

	t.Run("fast query completes successfully", func(t *testing.T) {
		e := echo.New()

		// Handler with fast query
		fastQueryHandler := func(c echo.Context) error {
			ctx := c.Request().Context()

			var result int
			err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, map[string]interface{}{"result": result})
		}

		middleware := TimeoutMiddleware(5*time.Second, logger)
		handler := middleware(fastQueryHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		echoCtx := e.NewContext(req, rec)

		err := handler(echoCtx)
		require.NoError(t, err)

		// Should complete successfully
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(1), response["result"])
	})

	t.Run("multiple queries cancelled on timeout", func(t *testing.T) {
		e := echo.New()

		queryExecuted := 0

		// Handler that tries multiple queries
		multiQueryHandler := func(c echo.Context) error {
			ctx := c.Request().Context()

			// First query (short, should succeed)
			var result1 int
			err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result1).Error
			if err != nil {
				return err
			}
			queryExecuted++

			// Second query (long, should timeout)
			var result2 int
			err = db.WithContext(ctx).Raw("SELECT pg_sleep(5)").Scan(&result2).Error
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			}
			queryExecuted++

			// Third query (should never execute)
			var result3 int
			err = db.WithContext(ctx).Raw("SELECT 3").Scan(&result3).Error
			if err != nil {
				return err
			}
			queryExecuted++

			return c.String(http.StatusOK, "all queries completed")
		}

		middleware := TimeoutMiddleware(100*time.Millisecond, logger)
		handler := middleware(multiQueryHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		echoCtx := e.NewContext(req, rec)

		err := handler(echoCtx)
		require.NoError(t, err)

		// Should timeout
		assert.Equal(t, http.StatusRequestTimeout, rec.Code)

		// Give goroutine time to complete
		time.Sleep(200 * time.Millisecond)

		// Only first query should have executed
		assert.Equal(t, 1, queryExecuted, "only first query should have completed before timeout")
	})
}

func TestTimeoutMiddleware_Integration_ContextPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// ctx := context.Background()

	connStr := helpers.GetTestDB(t)
	dbConn, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	defer dbConn.Close()

	db, err := gorm.Open(pgGormDriver.New(pgGormDriver.Config{
		Conn: dbConn,
	}), &gorm.Config{
		TranslateError: true,
	})
	require.NoError(t, err)

	logger, _ := logging.NewDevelopmentLogger()

	t.Run("context deadline propagates to database operations", func(t *testing.T) {
		e := echo.New()

		deadlineChecked := false

		handler := func(c echo.Context) error {
			ctx := c.Request().Context()

			// Verify context has deadline
			_, ok := ctx.Deadline()
			if !ok {
				return fmt.Errorf("context should have deadline")
			}
			deadlineChecked = true

			// Execute query with context
			var result int
			err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
		}

		middleware := TimeoutMiddleware(5*time.Second, logger)
		wrappedHandler := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		echoCtx := e.NewContext(req, rec)

		err := wrappedHandler(echoCtx)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, deadlineChecked, "deadline should have been checked")
	})
}
