package traveller

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

type travellerRepository struct {
	db     *gorm.DB
	logger *logging.Logger
}

func NewTravellerRepository(db *gorm.DB, logger *logging.Logger) *travellerRepository {
	return &travellerRepository{
		db:     db,
		logger: logger.Named("repository.traveller"),
	}
}
func (r *travellerRepository) GetByID(ctx context.Context, id int) (result domain.Traveller, err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.GetByID", "select", "m_traveller",
		attribute.Int("traveller.id", id),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	err = r.db.WithContext(ctx).Preload("Accessory").First(&result, "id = ?", id).Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("select", "m_traveller", duration),
		zap.Int("traveller.id", id),
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.WithContext(ctx).Warn("traveller not found", logFields...)
			return result, domain.NewNotFoundError("traveller", id)
		}
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to get traveller", logFields...)
		return
	}

	r.logger.WithContext(ctx).Debug("traveller retrieved",
		append(logFields,
			zap.String("traveller.name", result.Name),
			zap.Int("traveller.rarity", result.Rarity),
		)...)

	return
}

func (r *travellerRepository) GetList(ctx context.Context, filter domain.ListTravellerRequest, offset, limit int) (result []domain.Traveller, total int64, err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.GetList", "select", "m_traveller")
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	query := r.db.WithContext(ctx).Preload("Accessory")

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

func (r *travellerRepository) Create(ctx context.Context, input *domain.Traveller) (err error) {
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
		// Check for duplicate key violation
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			r.logger.WithContext(ctx).Warn("duplicate traveller name", append(logFields, logging.ErrorFields(err)...)...)
			return domain.NewConflictError("traveller with this name already exists")
		}
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to create traveller", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller created successfully",
		append(logFields, zap.Int64("traveller.id", input.ID))...)

	return
}

func (r *travellerRepository) Update(ctx context.Context, input *domain.Traveller) (err error) {
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

	result := r.db.WithContext(ctx).Updates(input)
	err = result.Error

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("update", "m_traveller", duration),
		zap.Int64("traveller.id", input.ID),
	)

	if err != nil {
		// Check for duplicate key violation
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			r.logger.WithContext(ctx).Warn("duplicate traveller name", append(logFields, logging.ErrorFields(err)...)...)
			return domain.NewConflictError("traveller with this name already exists")
		}
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("failed to update traveller", logFields...)
		return
	}

	// Check if any rows were affected (resource existed)
	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("traveller not found for update", logFields...)
		return domain.NewNotFoundError("traveller", input.ID)
	}

	r.logger.WithContext(ctx).Info("traveller updated successfully", logFields...)

	return
}

func (r *travellerRepository) Delete(ctx context.Context, id int) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.Delete", "delete", "m_traveller",
		attribute.Int("traveller.id", id),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("deleting traveller",
		zap.Int("traveller.id", id),
	)

	result := r.db.WithContext(ctx).Delete(&domain.Traveller{}, id)
	err = result.Error

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

	// Check if any rows were affected (resource existed)
	if result.RowsAffected == 0 {
		r.logger.WithContext(ctx).Warn("traveller not found for deletion", logFields...)
		return domain.NewNotFoundError("traveller", id)
	}

	r.logger.WithContext(ctx).Info("traveller deleted successfully", logFields...)

	return
}

// CreateTravellerWithAccessory creates a traveller and optionally an accessory in a single transaction
func (r *travellerRepository) CreateTravellerWithAccessory(ctx context.Context, traveller *domain.Traveller, accessory *domain.Accessory) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.CreateTravellerWithAccessory", "transaction", "m_traveller",
		attribute.String("traveller.name", traveller.Name),
		attribute.Bool("has_accessory", accessory != nil),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("creating traveller with accessory in transaction",
		zap.String("traveller.name", traveller.Name),
		zap.Bool("has_accessory", accessory != nil),
	)

	// Start transaction
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create accessory first if provided
		if accessory != nil {
			r.logger.WithContext(ctx).Info("creating accessory in transaction",
				zap.String("accessory.name", accessory.Name),
			)

			if err := tx.Create(accessory).Error; err != nil {
				r.logger.WithContext(ctx).Error("failed to create accessory in transaction",
					zap.String("accessory.name", accessory.Name),
					zap.Error(err),
				)
				return err
			}

			// Set accessory ID on traveller
			accessoryIDInt := int(accessory.ID)
			traveller.AccessoryID = &accessoryIDInt

			r.logger.WithContext(ctx).Info("accessory created in transaction",
				zap.Int64("accessory.id", accessory.ID),
			)
		}

		// Create traveller
		if err := tx.Create(traveller).Error; err != nil {
			// Check for duplicate key violation
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				r.logger.WithContext(ctx).Warn("duplicate traveller name",
					zap.String("traveller.name", traveller.Name),
					zap.Error(err),
				)
				return domain.NewConflictError("traveller with this name already exists")
			}
			r.logger.WithContext(ctx).Error("failed to create traveller in transaction",
				zap.String("traveller.name", traveller.Name),
				zap.Error(err),
			)
			return err
		}

		r.logger.WithContext(ctx).Info("traveller created in transaction",
			zap.Int64("traveller.id", traveller.ID),
		)

		return nil
	})

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("transaction", "m_traveller", duration),
		zap.String("traveller.name", traveller.Name),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("transaction failed", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller with accessory created successfully",
		append(logFields, zap.Int64("traveller.id", traveller.ID))...)

	return
}

// UpdateTravellerWithAccessory updates a traveller and handles accessory create/update in a single transaction
func (r *travellerRepository) UpdateTravellerWithAccessory(ctx context.Context, id int, traveller *domain.Traveller, accessory *domain.Accessory) (err error) {
	ctx, span := telemetry.StartDBSpan(ctx, "repository.traveller", "TravellerRepository.UpdateTravellerWithAccessory", "transaction", "m_traveller",
		attribute.Int("traveller.id", id),
		attribute.String("traveller.name", traveller.Name),
		attribute.Bool("has_accessory", accessory != nil),
	)
	defer telemetry.EndSpanWithError(span, err)

	start := time.Now()

	r.logger.WithContext(ctx).Info("updating traveller with accessory in transaction",
		zap.Int("traveller.id", id),
		zap.String("traveller.name", traveller.Name),
		zap.Bool("has_accessory", accessory != nil),
	)

	// Start transaction
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, fetch existing traveller to check if it has an accessory
		var existingTraveller domain.Traveller
		if err := tx.Select("id", "accessory_id").First(&existingTraveller, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.WithContext(ctx).Warn("traveller not found for update",
					zap.Int("traveller.id", id),
				)
				return domain.NewNotFoundError("traveller", id)
			}
			r.logger.WithContext(ctx).Error("failed to fetch existing traveller",
				zap.Int("traveller.id", id),
				zap.Error(err),
			)
			return err
		}

		// Handle accessory if provided
		if accessory != nil {
			if existingTraveller.AccessoryID != nil {
				// Update existing accessory
				accessory.ID = int64(*existingTraveller.AccessoryID)

				r.logger.WithContext(ctx).Info("updating existing accessory in transaction",
					zap.Int64("accessory.id", accessory.ID),
					zap.String("accessory.name", accessory.Name),
				)

				updateData := map[string]interface{}{
					"name":   accessory.Name,
					"hp":     accessory.HP,
					"sp":     accessory.SP,
					"patk":   accessory.PAtk,
					"pdef":   accessory.PDef,
					"eatk":   accessory.EAtk,
					"edef":   accessory.EDef,
					"spd":    accessory.Spd,
					"crit":   accessory.Crit,
					"effect": accessory.Effect,
				}
				if err := tx.Model(&domain.Accessory{}).Where("id = ?", accessory.ID).Updates(updateData).Error; err != nil {
					r.logger.WithContext(ctx).Error("failed to update accessory in transaction",
						zap.Int64("accessory.id", accessory.ID),
						zap.Error(err),
					)
					return err
				}

				traveller.AccessoryID = existingTraveller.AccessoryID

				r.logger.WithContext(ctx).Info("accessory updated in transaction",
					zap.Int64("accessory.id", accessory.ID),
				)
			} else {
				// Create new accessory
				r.logger.WithContext(ctx).Info("creating new accessory in transaction",
					zap.String("accessory.name", accessory.Name),
				)

				if err := tx.Create(accessory).Error; err != nil {
					r.logger.WithContext(ctx).Error("failed to create accessory in transaction",
						zap.String("accessory.name", accessory.Name),
						zap.Error(err),
					)
					return err
				}

				// Set new accessory ID on traveller
				accessoryIDInt := int(accessory.ID)
				traveller.AccessoryID = &accessoryIDInt

				r.logger.WithContext(ctx).Info("new accessory created in transaction",
					zap.Int64("accessory.id", accessory.ID),
				)
			}
		} else {
			// Keep existing accessory ID (no change to accessory)
			traveller.AccessoryID = existingTraveller.AccessoryID
		}

		// Update traveller
		result := tx.Updates(traveller)
		if err := result.Error; err != nil {
			// Check for duplicate key violation
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				r.logger.WithContext(ctx).Warn("duplicate traveller name",
					zap.String("traveller.name", traveller.Name),
					zap.Error(err),
				)
				return domain.NewConflictError("traveller with this name already exists")
			}
			r.logger.WithContext(ctx).Error("failed to update traveller in transaction",
				zap.Int("traveller.id", id),
				zap.Error(err),
			)
			return err
		}

		r.logger.WithContext(ctx).Info("traveller updated in transaction",
			zap.Int("traveller.id", id),
		)

		return nil
	})

	duration := time.Since(start)
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))
	logFields := append(
		logging.DatabaseFields("transaction", "m_traveller", duration),
		zap.Int("traveller.id", id),
	)

	if err != nil {
		logFields = append(logFields, logging.ErrorFields(err)...)
		r.logger.WithContext(ctx).Error("transaction failed", logFields...)
		return
	}

	r.logger.WithContext(ctx).Info("traveller with accessory updated successfully", logFields...)

	return
}
