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
	start := time.Now()

	err = r.db.WithContext(ctx).Preload("Influence").First(&result, "id = ?", id).Error

	duration := time.Since(start)
	logFields := []zap.Field{
		zap.Int("traveller.id", id),
		zap.String("db.system", "postgres"),
		zap.String("db.operation", "select"),
		zap.String("db.table", "tr_traveller"),
		zap.Float64("db.duration_ms", float64(duration.Milliseconds())),
	}

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

func (r TravellerRepository) Create(ctx context.Context, input *domain.Traveller) (err error) {
	start := time.Now()

	r.logger.WithContext(ctx).Info("creating traveller",
		zap.String("traveller.name", input.Name),
		zap.Int("traveller.rarity", input.Rarity),
		zap.Int("influence.id", int(input.InfluenceID)),
	)

	err = r.db.WithContext(ctx).Create(input).Error

	duration := time.Since(start)
	logFields := []zap.Field{
		zap.String("traveller.name", input.Name),
		zap.String("db.system", "postgres"),
		zap.String("db.operation", "insert"),
		zap.String("db.table", "tr_traveller"),
		zap.Float64("db.duration_ms", float64(duration.Milliseconds())),
	}

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
	start := time.Now()

	r.logger.WithContext(ctx).Info("updating traveller",
		zap.Int64("traveller.id", input.ID),
		zap.String("traveller.name", input.Name),
	)

	err = r.db.WithContext(ctx).Updates(input).Error

	duration := time.Since(start)
	logFields := []zap.Field{
		zap.Int64("traveller.id", input.ID),
		zap.String("db.system", "postgres"),
		zap.String("db.operation", "update"),
		zap.String("db.table", "tr_traveller"),
		zap.Float64("db.duration_ms", float64(duration.Milliseconds())),
	}

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to update traveller", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller updated successfully", logFields...)

	return
}

func (r TravellerRepository) Delete(ctx context.Context, id int) (err error) {
	start := time.Now()

	r.logger.WithContext(ctx).Info("deleting traveller",
		zap.Int("traveller.id", id),
	)

	err = r.db.WithContext(ctx).Delete(&domain.Traveller{}, id).Error

	duration := time.Since(start)
	logFields := []zap.Field{
		zap.Int("traveller.id", id),
		zap.String("db.system", "postgres"),
		zap.String("db.operation", "delete"),
		zap.String("db.table", "tr_traveller"),
		zap.Float64("db.duration_ms", float64(duration.Milliseconds())),
	}

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to delete traveller", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller deleted successfully", logFields...)

	return
}
