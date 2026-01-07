package logging

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type FieldsSuite struct {
	suite.Suite
}

func TestFieldsSuite(t *testing.T) {
	suite.Run(t, new(FieldsSuite))
}

func (s *FieldsSuite) TestFields_HTTPFields() {
	type args struct {
		method     string
		route      string
		statusCode int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "GET request",
			args: args{
				method:     "GET",
				route:      "/api/v1/travellers/:id",
				statusCode: 200,
			},
		},
		{
			name: "POST request",
			args: args{
				method:     "POST",
				route:      "/api/v1/travellers",
				statusCode: 201,
			},
		},
		{
			name: "error response",
			args: args{
				method:     "PUT",
				route:      "/api/v1/travellers/:id",
				statusCode: 400,
			},
		},
		{
			name: "server error",
			args: args{
				method:     "DELETE",
				route:      "/api/v1/travellers/:id",
				statusCode: 500,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fields := HTTPFields(tt.args.method, tt.args.route, tt.args.statusCode)

			assert.NotNil(s.T(), fields)
			assert.Len(s.T(), fields, 3)
			assert.Equal(s.T(), zap.String("http.method", tt.args.method), fields[0])
			assert.Equal(s.T(), zap.String("http.route", tt.args.route), fields[1])
			assert.Equal(s.T(), zap.Int("http.status_code", tt.args.statusCode), fields[2])
		})
	}
}

func (s *FieldsSuite) TestFields_ErrorFields() {
	type args struct {
		err error
	}
	tests := []struct {
		name          string
		args          args
		expectedCount int
	}{
		{
			name: "with error",
			args: args{
				err: errors.New("test error message"),
			},
			expectedCount: 2,
		},
		{
			name: "nil error",
			args: args{
				err: nil,
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fields := ErrorFields(tt.args.err)

			assert.NotNil(s.T(), fields)
			assert.Len(s.T(), fields, tt.expectedCount)

			if tt.args.err != nil {
				assert.Equal(s.T(), zap.String("error.message", tt.args.err.Error()), fields[0])
				assert.Equal(s.T(), zap.String("error.type", "error"), fields[1])
			}
		})
	}
}

func (s *FieldsSuite) TestFields_DatabaseFields() {
	type args struct {
		operation string
		table     string
		duration  time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "select operation",
			args: args{
				operation: "select",
				table:     "tr_traveller",
				duration:  15 * time.Millisecond,
			},
		},
		{
			name: "insert operation",
			args: args{
				operation: "insert",
				table:     "tr_traveller",
				duration:  25 * time.Millisecond,
			},
		},
		{
			name: "update operation",
			args: args{
				operation: "update",
				table:     "tr_traveller",
				duration:  20 * time.Millisecond,
			},
		},
		{
			name: "delete operation",
			args: args{
				operation: "delete",
				table:     "tr_traveller",
				duration:  10 * time.Millisecond,
			},
		},
		{
			name: "long duration query",
			args: args{
				operation: "select",
				table:     "m_user",
				duration:  500 * time.Millisecond,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fields := DatabaseFields(tt.args.operation, tt.args.table, tt.args.duration)

			assert.NotNil(s.T(), fields)
			assert.Len(s.T(), fields, 4)
			assert.Equal(s.T(), zap.String("db.system", "postgres"), fields[0])
			assert.Equal(s.T(), zap.String("db.operation", tt.args.operation), fields[1])
			assert.Equal(s.T(), zap.String("db.table", tt.args.table), fields[2])
			assert.Equal(s.T(), zap.Float64("db.duration_ms", float64(tt.args.duration.Milliseconds())), fields[3])
		})
	}
}

func (s *FieldsSuite) TestFields_UserFields() {
	type args struct {
		userID   string
		username string
	}
	tests := []struct {
		name          string
		args          args
		expectedCount int
	}{
		{
			name: "with both user ID and username",
			args: args{
				userID:   "user-123",
				username: "john_doe",
			},
			expectedCount: 2,
		},
		{
			name: "with user ID only",
			args: args{
				userID:   "user-456",
				username: "",
			},
			expectedCount: 1,
		},
		{
			name: "with username only",
			args: args{
				userID:   "",
				username: "jane_smith",
			},
			expectedCount: 1,
		},
		{
			name: "empty user fields",
			args: args{
				userID:   "",
				username: "",
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fields := UserFields(tt.args.userID, tt.args.username)

			assert.NotNil(s.T(), fields)
			assert.Len(s.T(), fields, tt.expectedCount)

			// Check field contents based on what was provided
			fieldIndex := 0
			if tt.args.userID != "" {
				assert.Equal(s.T(), zap.String("user.id", tt.args.userID), fields[fieldIndex])
				fieldIndex++
			}
			if tt.args.username != "" {
				assert.Equal(s.T(), zap.String("user.username", tt.args.username), fields[fieldIndex])
			}
		})
	}
}

func (s *FieldsSuite) TestFields_TraceFields() {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name          string
		args          args
		expectedCount int
	}{
		{
			name: "empty context - OTel not integrated",
			args: args{
				ctx: context.Background(),
			},
			expectedCount: 0,
		},
		{
			name: "context with values - OTel not integrated",
			args: args{
				ctx: WithRequestID(context.Background(), "test-request-id"),
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			fields := TraceFields(tt.args.ctx)

			assert.NotNil(s.T(), fields)
			assert.Len(s.T(), fields, tt.expectedCount)
			// Currently should be empty since OTel is not integrated
			// When OTel is added, this test will need to be updated
		})
	}
}

func (s *FieldsSuite) TestFields_Integration() {
	s.Run("test all field helpers together", func() {
		logger, err := NewDevelopmentLogger()
		assert.Nil(s.T(), err)
		defer logger.Sync()

		// Simulate a complete request with all field types
		ctx := context.Background()
		ctx = WithRequestID(ctx, "req-123")
		ctx = WithUserID(ctx, "user-456")

		httpFields := HTTPFields("GET", "/api/v1/travellers/1", 200)
		dbFields := DatabaseFields("select", "tr_traveller", 25*time.Millisecond)
		userFields := UserFields("user-456", "testuser")
		traceFields := TraceFields(ctx)

		// Log with all fields
		allFields := append(httpFields, dbFields...)
		allFields = append(allFields, userFields...)
		allFields = append(allFields, traceFields...)

		ctxLogger := logger.WithContext(ctx)
		ctxLogger.Info("test integration message", allFields...)
	})
}
