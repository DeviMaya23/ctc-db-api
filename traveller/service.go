package traveller

import (
	"context"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type TravellerRepository interface {
	GetByID(ctx context.Context, id int) (result domain.Traveller, err error)
	Create(ctx context.Context, input *domain.Traveller) (err error)
	Update(ctx context.Context, input *domain.Traveller) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type AccessoryRepository interface {
	Create(ctx context.Context, input *domain.Accessory) (err error)
	Update(ctx context.Context, input *domain.Accessory) (err error)
}

type Service struct {
	travellerRepo  TravellerRepository
	accessoryRepo  AccessoryRepository
	logger         *logging.Logger
}

func NewTravellerService(t TravellerRepository, a AccessoryRepository, logger *logging.Logger) *Service {
	return &Service{
		travellerRepo:  t,
		accessoryRepo:  a,
		logger:         logger.Named("service.traveller"),
	}
}

func (s Service) GetByID(ctx context.Context, id int) (res domain.Traveller, err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.traveller", "TravellerService.GetByID",
		attribute.Int("traveller.id", id),
	)
	defer telemetry.EndSpanWithError(span, err)

	s.logger.WithContext(ctx).Info("fetching traveller",
		zap.Int("traveller.id", id),
	)

	res, err = s.travellerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to fetch traveller",
			zap.Int("traveller.id", id),
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	s.logger.WithContext(ctx).Info("traveller fetched successfully",
		zap.Int("traveller.id", id),
		zap.String("traveller.name", res.Name),
	)

	return
}

func (s Service) Create(ctx context.Context, input domain.CreateTravellerRequest) (err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.traveller", "TravellerService.Create",
		attribute.String("traveller.name", input.Name),
	)
	defer telemetry.EndSpanWithError(span, err)

	s.logger.WithContext(ctx).Info("creating traveller",
		zap.String("traveller.name", input.Name),
	)

	// Create accessory if provided
	var accessoryID *int
	if input.Accessory != nil {
		s.logger.WithContext(ctx).Info("creating accessory for traveller",
			zap.String("accessory.name", input.Accessory.Name),
		)

		newAccessory := domain.Accessory{
			Name:   input.Accessory.Name,
			HP:     input.Accessory.HP,
			SP:     input.Accessory.SP,
			PAtk:   input.Accessory.PAtk,
			PDef:   input.Accessory.PDef,
			EAtk:   input.Accessory.EAtk,
			EDef:   input.Accessory.EDef,
			Spd:    input.Accessory.Spd,
			Crit:   input.Accessory.Crit,
			Effect: input.Accessory.Effect,
		}

		err = s.accessoryRepo.Create(ctx, &newAccessory)
		if err != nil {
			s.logger.WithContext(ctx).Error("failed to create accessory",
				zap.String("accessory.name", input.Accessory.Name),
				zap.String("error.type", "repository_error"),
				zap.String("error.message", err.Error()),
			)
			return
		}

		accessoryID = new(int)
		*accessoryID = int(newAccessory.ID)

		s.logger.WithContext(ctx).Info("accessory created successfully",
			zap.String("accessory.name", input.Accessory.Name),
			zap.Int64("accessory.id", newAccessory.ID),
		)
	}

	newTraveller := domain.Traveller{
		Name:        input.Name,
		Rarity:      input.Rarity,
		InfluenceID: constants.GetInfluenceID(input.Influence),
		AccessoryID: accessoryID,
	}

	err = s.travellerRepo.Create(ctx, &newTraveller)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to create traveller",
			zap.String("traveller.name", input.Name),
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	s.logger.WithContext(ctx).Info("traveller created successfully",
		zap.String("traveller.name", input.Name),
		zap.Int64("traveller.id", newTraveller.ID),
	)

	return
}

func (s Service) Update(ctx context.Context, id int, input domain.UpdateTravellerRequest) (err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.traveller", "TravellerService.Update",
		attribute.Int("traveller.id", id),
		attribute.String("traveller.name", input.Name),
	)
	defer telemetry.EndSpanWithError(span, err)

	s.logger.WithContext(ctx).Info("updating traveller",
		zap.Int("traveller.id", id),
		zap.String("traveller.name", input.Name),
	)

	// First, get the existing traveller to check if it has an accessory
	existingTraveller, err := s.travellerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to fetch existing traveller",
			zap.Int("traveller.id", id),
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	// Handle accessory update/creation
	var accessoryID *int
	if input.Accessory != nil {
		if existingTraveller.AccessoryID != nil {
			// Update existing accessory
			s.logger.WithContext(ctx).Info("updating existing accessory",
				zap.Int("accessory.id", *existingTraveller.AccessoryID),
				zap.String("accessory.name", input.Accessory.Name),
			)

			updatedAccessory := domain.Accessory{
				CommonModel: domain.CommonModel{ID: int64(*existingTraveller.AccessoryID)},
				Name:        input.Accessory.Name,
				HP:          input.Accessory.HP,
				SP:          input.Accessory.SP,
				PAtk:        input.Accessory.PAtk,
				PDef:        input.Accessory.PDef,
				EAtk:        input.Accessory.EAtk,
				EDef:        input.Accessory.EDef,
				Spd:         input.Accessory.Spd,
				Crit:        input.Accessory.Crit,
				Effect:      input.Accessory.Effect,
			}

			err = s.accessoryRepo.Update(ctx, &updatedAccessory)
			if err != nil {
				s.logger.WithContext(ctx).Error("failed to update accessory",
					zap.Int("accessory.id", *existingTraveller.AccessoryID),
					zap.String("error.type", "repository_error"),
					zap.String("error.message", err.Error()),
				)
				return
			}

			accessoryID = existingTraveller.AccessoryID
			s.logger.WithContext(ctx).Info("accessory updated successfully",
				zap.Int("accessory.id", *existingTraveller.AccessoryID),
			)
		} else {
			// Create new accessory
			s.logger.WithContext(ctx).Info("creating new accessory for traveller",
				zap.String("accessory.name", input.Accessory.Name),
			)

			newAccessory := domain.Accessory{
				Name:   input.Accessory.Name,
				HP:     input.Accessory.HP,
				SP:     input.Accessory.SP,
				PAtk:   input.Accessory.PAtk,
				PDef:   input.Accessory.PDef,
				EAtk:   input.Accessory.EAtk,
				EDef:   input.Accessory.EDef,
				Spd:    input.Accessory.Spd,
				Crit:   input.Accessory.Crit,
				Effect: input.Accessory.Effect,
			}

			err = s.accessoryRepo.Create(ctx, &newAccessory)
			if err != nil {
				s.logger.WithContext(ctx).Error("failed to create accessory",
					zap.String("accessory.name", input.Accessory.Name),
					zap.String("error.type", "repository_error"),
					zap.String("error.message", err.Error()),
				)
				return
			}

			accessoryID = new(int)
			*accessoryID = int(newAccessory.ID)

			s.logger.WithContext(ctx).Info("accessory created successfully",
				zap.String("accessory.name", input.Accessory.Name),
				zap.Int64("accessory.id", newAccessory.ID),
			)
		}
	} else {
		// Keep existing accessory ID (no change)
		accessoryID = existingTraveller.AccessoryID
	}

	// Update traveller
	updatedTraveller := domain.Traveller{
		CommonModel: domain.CommonModel{ID: int64(id)},
		Name:        input.Name,
		Rarity:      input.Rarity,
		InfluenceID: constants.GetInfluenceID(input.Influence),
		AccessoryID: accessoryID,
	}

	err = s.travellerRepo.Update(ctx, &updatedTraveller)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to update traveller",
			zap.Int("traveller.id", id),
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	s.logger.WithContext(ctx).Info("traveller updated successfully",
		zap.Int("traveller.id", id),
	)

	return
}

func (s Service) Delete(ctx context.Context, id int) (err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.traveller", "TravellerService.Delete",
		attribute.Int("traveller.id", id),
	)
	defer telemetry.EndSpanWithError(span, err)

	s.logger.WithContext(ctx).Info("deleting traveller",
		zap.Int("traveller.id", id),
	)

	err = s.travellerRepo.Delete(ctx, id)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to delete traveller",
			zap.Int("traveller.id", id),
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	s.logger.WithContext(ctx).Info("traveller deleted successfully",
		zap.Int("traveller.id", id),
	)

	return
}
