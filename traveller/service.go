package traveller

import (
	"context"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"

	"go.uber.org/zap"
)

type TravellerRepository interface {
	GetByID(ctx context.Context, id int) (result domain.Traveller, err error)
	Create(ctx context.Context, input *domain.Traveller) (err error)
	Update(ctx context.Context, input *domain.Traveller) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type Service struct {
	travellerRepo TravellerRepository
	logger        *logging.Logger
}

func NewTravellerService(t TravellerRepository, logger *logging.Logger) *Service {
	return &Service{
		travellerRepo: t,
		logger:        logger.Named("service.traveller"),
	}
}

func (s Service) GetByID(ctx context.Context, id int) (res domain.Traveller, err error) {
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
		zap.Int("traveller.rarity", res.Rarity),
	)

	return
}

func (s Service) Create(ctx context.Context, input domain.CreateTravellerRequest) (err error) {
	s.logger.WithContext(ctx).Info("creating traveller",
		zap.String("traveller.name", input.Name),
		zap.String("influence.name", input.Influence),
		zap.Int("traveller.rarity", input.Rarity),
	)

	newTraveller := domain.Traveller{
		Name:        input.Name,
		Rarity:      input.Rarity,
		InfluenceID: constants.GetInfluenceID(input.Influence),
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

func (s Service) Update(ctx context.Context, input *domain.Traveller) (err error) {
	s.logger.WithContext(ctx).Info("updating traveller",
		zap.Int64("traveller.id", input.ID),
		zap.String("traveller.name", input.Name),
	)

	err = s.travellerRepo.Update(ctx, input)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to update traveller",
			zap.Int64("traveller.id", input.ID),
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	s.logger.WithContext(ctx).Info("traveller updated successfully",
		zap.Int64("traveller.id", input.ID),
	)

	return
}

func (s Service) Delete(ctx context.Context, id int) (err error) {
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
