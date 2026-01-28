package user

import (
	"context"
	"errors"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userRepository struct {
	db     *gorm.DB
	logger *logging.Logger
}

func NewUserRepository(db *gorm.DB, logger *logging.Logger) *userRepository {
	return &userRepository{
		db:     db,
		logger: logger.Named("repository.user"),
	}
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (result *domain.User, err error) {
	// Start database span
	ctx, op := telemetry.StartDBSpan(ctx, "repository.user", "UserRepository.GetByUsername", "select", "m_user",
		attribute.String("user.username", username),
	)
	defer op.End(err)

	result = &domain.User{}
	err = r.db.WithContext(ctx).First(result, "username = ?", username).Error

	logFields := append(
		logging.DatabaseFields("select", "m_user", op.Duration()),
		zap.String("user.username", username),
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("user not found", logFields...)
			return nil, domain.NewNotFoundError("user", username)
		}
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to get user", logFields...)
		return
	}

	return
}
