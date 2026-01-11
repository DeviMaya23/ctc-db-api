package postgres

import (
	"context"
	"errors"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TravellerRepository struct {
	db     *gorm.DB
	logger *logging.Logger
}

func NewTravellerRepository(db *gorm.DB, logger *logging.Logger) *TravellerRepository {
	return &TravellerRepository{
		db:     db,
		logger: logger.Named("repository.traveller"),
	}
}
func (r TravellerRepository) GetByID(ctx context.Context, id int) (result domain.Traveller, err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.GetByID", "select", "m_traveller",
		attribute.Int("traveller.id", id),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	err = r.db.WithContext(ctx).Preload("Influence").Preload("Job").Preload("Accessory").First(&result, "id = ?", id).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("select", "m_traveller", duration),
		zap.Int("traveller.id", id),
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("traveller not found", logFields...)
		} else {
			logFields = append(logFields, logging.ErrorFields(err)...)
			r.logger.WithContext(ctx).Error("failed to get traveller", logFields...)
		}
		return
	}

	r.logger.WithContext(ctx).Debug("traveller retrieved",
		append(logFields,
			zap.String("traveller.name", result.Name),
			zap.Int("traveller.rarity", result.Rarity),
		)...)

	return
}

func (r TravellerRepository) GetList(ctx context.Context, filter domain.ListTravellerRequest, offset, limit int) (result []domain.Traveller, total int64, err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.GetList", "select", "m_traveller")
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	query := r.db.WithContext(ctx).Preload("Influence").Preload("Job").Preload("Accessory")

	// Apply filters
	if filter.Name != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+filter.Name+"%")
	}
	if filter.InfluenceID != 0 {
		query = query.Where("influence_id = ?", filter.InfluenceID)
	}
	if filter.JobID != 0 {
		query = query.Where("job_id = ?", filter.JobID)
	}

	// Get total count
	err = query.Model(&domain.Traveller{}).Count(&total).Error
	if err != nil {
		r.logger.WithContext(ctx).Error("failed to count travellers", zap.Error(err))
		return
	}

	// Apply pagination
	err = query.Offset(offset).Limit(limit).Find(&result).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("select", "m_traveller", duration),
		zap.Int64("total", total),
		zap.Int("returned", len(result)),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to get traveller list", logFields...)
		return
	}

	r.logger.WithContext(ctx).Debug("traveller list retrieved", logFields...)

	return
}

func (r TravellerRepository) Create(ctx context.Context, input *domain.Traveller) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.Create", "insert", "m_traveller",
		attribute.String("traveller.name", input.Name),
		attribute.Int("traveller.rarity", input.Rarity),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("creating traveller",
		zap.String("traveller.name", input.Name),
		zap.Int("traveller.rarity", input.Rarity),
		zap.Int("influence.id", int(input.InfluenceID)),
		zap.Int("job.id", int(input.JobID)),
	)

	err = r.db.WithContext(ctx).Create(input).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("insert", "m_traveller", duration),
		zap.String("traveller.name", input.Name),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to create traveller", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller created successfully",
		append(logFields, zap.Int64("traveller.id", input.ID))...)

	return
}

func (r TravellerRepository) Update(ctx context.Context, input *domain.Traveller) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.Update", "update", "m_traveller",
		attribute.Int64("traveller.id", input.ID),
		attribute.String("traveller.name", input.Name),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("updating traveller",
		zap.Int64("traveller.id", input.ID),
		zap.String("traveller.name", input.Name),
	)

	err = r.db.WithContext(ctx).Updates(input).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("update", "m_traveller", duration),
		zap.Int64("traveller.id", input.ID),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to update traveller", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller updated successfully", logFields...)

	return
}

func (r TravellerRepository) Delete(ctx context.Context, id int) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.Delete", "delete", "m_traveller",
		attribute.Int("traveller.id", id),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("deleting traveller",
		zap.Int("traveller.id", id),
	)

	err = r.db.WithContext(ctx).Delete(&domain.Traveller{}, id).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("delete", "m_traveller", duration),
		zap.Int("traveller.id", id),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to delete traveller", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller deleted successfully", logFields...)

	return
}
