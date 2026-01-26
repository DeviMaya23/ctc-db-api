package helpers

import (
	"database/sql/driver"
	"encoding/json"
	"io"
	"lizobly/ctc-db-api/pkg/validator"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

type (
	AnyTime struct{}
)

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func GetHTTPTestRecorder(t *testing.T, method, url string, requestBody interface{}, queryParams url.Values, pathParam map[string]string) (*httptest.ResponseRecorder, echo.Context) {
	t.Helper()

	e := echo.New()

	validator := validator.NewValidator()
	e.Validator = validator

	var body io.Reader
	if requestBody != nil {
		requestBytes, err := json.Marshal(requestBody)
		assert.NoError(t, err)
		body = strings.NewReader(string(requestBytes))
	}

	req := httptest.NewRequest(method, url, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	if queryParams != nil {
		query := req.URL.Query()
		for key, values := range queryParams {
			for _, value := range values {
				query.Add(key, value)
			}
		}
		req.URL.RawQuery = query.Encode()
	}
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	for key, value := range pathParam {
		ctx.SetParamNames(key)
		ctx.SetParamValues(value)
	}

	ctx.Set("validator", validator)

	return rec, ctx

}
