package postgres

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AccessoryRepository struct {
	db     *gorm.DB
	logger *logging.Logger
}

func NewAccessoryRepository(db *gorm.DB, logger *logging.Logger) *AccessoryRepository {
	return &AccessoryRepository{
		db:     db,
		logger: logger.Named("repository.accessory"),
	}
}

func (r AccessoryRepository) Create(ctx context.Context, input *domain.Accessory) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.accessory", "AccessoryRepository.Create", "insert", "m_accessory",
		attribute.String("accessory.name", input.Name),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("creating accessory",
		zap.String("accessory.name", input.Name),
	)

	err = r.db.WithContext(ctx).Create(input).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("insert", "m_accessory", duration),
		zap.String("accessory.name", input.Name),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to create accessory", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("accessory created successfully",
		append(logFields, zap.Int64("accessory.id", input.ID))...)

	return
}

func (r AccessoryRepository) Update(ctx context.Context, input *domain.Accessory) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.accessory", "AccessoryRepository.Update", "update", "m_accessory",
		attribute.Int64("accessory.id", input.ID),
		attribute.String("accessory.name", input.Name),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("updating accessory",
		zap.Int64("accessory.id", input.ID),
		zap.String("accessory.name", input.Name),
	)

	err = r.db.WithContext(ctx).Updates(input).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("update", "m_accessory", duration),
		zap.Int64("accessory.id", input.ID),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to update accessory", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("accessory updated successfully", logFields...)

	return
}
