package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type ZapLoggerSuite struct {
	suite.Suite
}

func TestZapLoggerSuite(t *testing.T) {
	suite.Run(t, new(ZapLoggerSuite))
}

func (s *ZapLoggerSuite) TestLogger_NewDevelopmentLogger() {
	s.Run("success", func() {
		logger, err := NewDevelopmentLogger()
		assert.Nil(s.T(), err)
		assert.NotNil(s.T(), logger)
		defer logger.Sync()

		// Test basic logging
		logger.Info("test message", zap.String("test.field", "value"))
	})
}

func (s *ZapLoggerSuite) TestLogger_NewProductionLogger() {
	s.Run("success", func() {
		logger, err := NewProductionLogger()
		assert.Nil(s.T(), err)
		assert.NotNil(s.T(), logger)
		defer logger.Sync()

		// Test basic logging
		logger.Info("test message", zap.String("test.field", "value"))
	})
}

func (s *ZapLoggerSuite) TestLogger_NewLogger() {
	type args struct {
		env string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "development environment",
			args:    args{env: "development"},
			wantErr: false,
		},
		{
			name:    "production environment",
			args:    args{env: "production"},
			wantErr: false,
		},
		{
			name:    "unknown environment defaults to development",
			args:    args{env: "staging"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			logger, err := NewLogger(tt.args.env)

			if tt.wantErr {
				assert.NotNil(s.T(), err)
				return
			}

			assert.Nil(s.T(), err)
			assert.NotNil(s.T(), logger)
			defer logger.Sync()

			logger.Info("test message",
				zap.String("environment", tt.args.env),
			)
		})
	}
}

func (s *ZapLoggerSuite) TestLogger_WithContext() {
	type args struct {
		requestID string
		userID    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "with request ID and user ID",
			args: args{
				requestID: "test-request-123",
				userID:    "test-user",
			},
		},
		{
			name: "with request ID only",
			args: args{
				requestID: "test-request-456",
			},
		},
		{
			name: "with user ID only",
			args: args{
				userID: "another-user",
			},
		},
		{
			name: "empty context",
			args: args{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			logger, err := NewDevelopmentLogger()
			assert.Nil(s.T(), err)
			defer logger.Sync()

			// Create context with values
			ctx := context.Background()
			if tt.args.requestID != "" {
				ctx = WithRequestID(ctx, tt.args.requestID)
			}
			if tt.args.userID != "" {
				ctx = WithUserID(ctx, tt.args.userID)
			}

			// Log with context
			ctxLogger := logger.WithContext(ctx)
			assert.NotNil(s.T(), ctxLogger)
			ctxLogger.Info("test message with context")
		})
	}
}

func (s *ZapLoggerSuite) TestLogger_Named() {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "service logger",
			args: args{name: "service.test"},
		},
		{
			name: "repository logger",
			args: args{name: "repository.postgres"},
		},
		{
			name: "handler logger",
			args: args{name: "handler.traveller"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			logger, err := NewDevelopmentLogger()
			assert.Nil(s.T(), err)
			defer logger.Sync()

			// Create named logger
			namedLogger := logger.Named(tt.args.name)
			assert.NotNil(s.T(), namedLogger)
			namedLogger.Info("test message from named logger")
		})
	}
}

func (s *ZapLoggerSuite) TestLogger_ExtractTraceID() {
	s.Run("returns empty string when OTel not integrated", func() {
		ctx := context.Background()

		// Currently should return empty string (stubbed for OTel)
		traceID := ExtractTraceID(ctx)
		assert.Empty(s.T(), traceID)
	})
}

func (s *ZapLoggerSuite) TestLogger_ExtractSpanID() {
	s.Run("returns empty string when OTel not integrated", func() {
		ctx := context.Background()

		// Currently should return empty string (stubbed for OTel)
		spanID := ExtractSpanID(ctx)
		assert.Empty(s.T(), spanID)
	})
}
