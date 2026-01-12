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

func (r AccessoryRepository) GetList(ctx context.Context, filter domain.ListAccessoryRequest, offset, limit int) (result []domain.Accessory, ownerNames map[int64]string, total int64, err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.accessory", "AccessoryRepository.GetList", "select", "m_accessory")
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	query := r.db.WithContext(ctx).
		Model(&domain.Accessory{}).
		Select("m_accessory.*, m_traveller.name as owner").
		Joins("LEFT JOIN m_traveller ON m_accessory.id = m_traveller.accessory_id")

	if filter.Effect != "" {
		query = query.Where("LOWER(m_accessory.effect) LIKE LOWER(?)", "%"+filter.Effect+"%")
	}

	if filter.Owner != "" {
		query = query.Where("LOWER(m_traveller.name) LIKE LOWER(?)", "%"+filter.Owner+"%")
	}

	err = query.Count(&total).Error
	if err != nil {
		r.logger.WithContext(ctx).Error("failed to count accessories", zap.Error(err))
		return
	}

	// Apply ordering if specified
	if filter.OrderBy != "" {
		orderDir := "DESC"
		if filter.OrderDir != "" {
			orderDir = filter.OrderDir
		}
		// Prefix with table name for clarity
		query = query.Order("m_accessory." + filter.OrderBy + " " + orderDir)
	}

	// Fetch accessories with traveller names in one query
	var rows []struct {
		domain.Accessory
		Owner string
	}
	err = query.Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		r.logger.WithContext(ctx).Error("failed to get accessory list", zap.Error(err))
		return
	}

	// Separate accessories and traveller names from the result
	result = make([]domain.Accessory, len(rows))
	ownerNames = make(map[int64]string)
	for i, row := range rows {
		result[i] = row.Accessory
		if row.Owner != "" {
			ownerNames[row.Accessory.ID] = row.Owner
		}
	}

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("select", "m_accessory", duration),
		zap.Int64("total", total),
		zap.Int("returned", len(result)),
	)

	r.logger.WithContext(ctx).Debug("accessory list retrieved", logFields...)

	return
}
