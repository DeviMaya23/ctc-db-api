package postgres

import (
	"context"
	"errors"
	"lizobly/cotc-db-api/pkg/domain"
	"lizobly/cotc-db-api/pkg/logging"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository struct {
	db     *gorm.DB
	logger *logging.Logger
}

func NewUserRepository(db *gorm.DB, logger *logging.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.Named("repository.user"),
	}
}

func (r UserRepository) GetByUsername(ctx context.Context, username string) (result domain.User, err error) {
	start := time.Now()

	err = r.db.WithContext(ctx).First(&result, "username = ?", username).Error

	duration := time.Since(start)
	logFields := append(
		logging.DatabaseFields("select", "m_user", duration),
		zap.String("user.username", username),
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("user not found", logFields...)
		} else {
			logFields = append(logFields, logging.ErrorFields(err)...)
			r.logger.WithContext(ctx).Error("failed to get user", logFields...)
		}
		return
	}

	r.logger.WithContext(ctx).Debug("user retrieved",
		append(logFields, zap.Int64("user.id", result.ID))...)

	return
}
