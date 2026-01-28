package traveller

import (
	"context"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type TravellerRepository interface {
	GetByID(ctx context.Context, id int) (result *domain.Traveller, err error)
	GetList(ctx context.Context, filter domain.ListTravellerRequest, offset, limit int) (result []*domain.Traveller, total int64, err error)
	Create(ctx context.Context, input *domain.Traveller) (err error)
	Update(ctx context.Context, input *domain.Traveller) (err error)
	Delete(ctx context.Context, id int) (err error)
	CreateTravellerWithAccessory(ctx context.Context, traveller *domain.Traveller, accessory *domain.Accessory) (err error)
	UpdateTravellerWithAccessory(ctx context.Context, id int, traveller *domain.Traveller, accessory *domain.Accessory) (err error)
}

type travellerService struct {
	travellerRepo TravellerRepository
	logger        *logging.Logger
}

func NewTravellerService(t TravellerRepository, logger *logging.Logger) *travellerService {
	return &travellerService{
		travellerRepo: t,
		logger:        logger.Named("service.traveller"),
	}
}

func (s *travellerService) GetByID(ctx context.Context, id int) (res *domain.Traveller, err error) {
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

func (s *travellerService) GetList(ctx context.Context, filter domain.ListTravellerRequest, params helpers.PaginationParams) (res helpers.PaginatedResponse[domain.TravellerListItemResponse], err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.traveller", "TravellerService.GetList",
		attribute.Int("page", params.Page),
		attribute.Int("page_size", params.PageSize),
	)
	defer telemetry.EndSpanWithError(span, err)

	// Normalize pagination params
	params.Normalize()

	// Populate ID fields from plaintext values
	if filter.Influence != "" {
		filter.InfluenceID = constants.GetInfluenceID(filter.Influence)
	}
	if filter.Job != "" {
		filter.JobID = constants.GetJobID(filter.Job)
	}

	s.logger.WithContext(ctx).Info("fetching traveller list",
		zap.Int("page", params.Page),
		zap.Int("page_size", params.PageSize),
		zap.String("filter.name", filter.Name),
		zap.String("filter.influence", filter.Influence),
		zap.String("filter.job", filter.Job),
	)

	travellers, total, err := s.travellerRepo.GetList(ctx, filter, params.Offset(), params.PageSize)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to fetch traveller list",
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	s.logger.WithContext(ctx).Info("traveller list fetched successfully",
		zap.Int64("total", total),
		zap.Int("returned", len(travellers)),
	)

	// Map to response DTOs
	items := make([]domain.TravellerListItemResponse, len(travellers))
	for i, t := range travellers {
		items[i] = domain.ToTravellerListItemResponse(t)
	}

	res = helpers.NewPaginatedResponse(items, params, total)

	return
}

func (s *travellerService) Create(ctx context.Context, input domain.CreateTravellerRequest) (id int64, err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.traveller", "TravellerService.Create",
		attribute.String("traveller.name", input.Name),
	)
	defer telemetry.EndSpanWithError(span, err)

	s.logger.WithContext(ctx).Info("creating traveller",
		zap.String("traveller.name", input.Name),
	)

	// Parse release date
	releaseDate, err := helpers.ParseDate(input.ReleaseDate, constants.DateFormat)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to parse release date",
			zap.String("release_date", input.ReleaseDate),
			zap.String("error.type", "parsing_error"),
			zap.String("error.message", err.Error()),
		)
		return 0, err
	}

	// Build traveller domain object
	newTraveller := domain.Traveller{
		Name:        input.Name,
		Rarity:      input.Rarity,
		Banner:      input.Banner,
		ReleaseDate: releaseDate,
		InfluenceID: constants.GetInfluenceID(input.Influence),
		JobID:       constants.GetJobID(input.Job),
	}

	// Build accessory domain object if provided
	var newAccessory *domain.Accessory
	if input.Accessory != nil {
		newAccessory = &domain.Accessory{
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
	}

	// Create traveller with accessory in transaction
	err = s.travellerRepo.CreateTravellerWithAccessory(ctx, &newTraveller, newAccessory)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to create traveller",
			zap.String("traveller.name", input.Name),
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return 0, err
	}

	s.logger.WithContext(ctx).Info("traveller created successfully",
		zap.String("traveller.name", input.Name),
		zap.Int64("traveller.id", newTraveller.ID),
	)

	return newTraveller.ID, nil
}

func (s *travellerService) Update(ctx context.Context, id int, input domain.UpdateTravellerRequest) (err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.traveller", "TravellerService.Update",
		attribute.Int("traveller.id", id),
		attribute.String("traveller.name", input.Name),
	)
	defer telemetry.EndSpanWithError(span, err)

	s.logger.WithContext(ctx).Info("updating traveller",
		zap.Int("traveller.id", id),
		zap.String("traveller.name", input.Name),
	)

	// Parse release date
	releaseDate, err := helpers.ParseDate(input.ReleaseDate, constants.DateFormat)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to parse release date",
			zap.String("release_date", input.ReleaseDate),
			zap.String("error.type", "parsing_error"),
			zap.String("error.message", err.Error()),
		)
		return err
	}

	// Build traveller domain object
	updatedTraveller := domain.Traveller{
		CommonModel: domain.CommonModel{ID: int64(id)},
		Name:        input.Name,
		Rarity:      input.Rarity,
		Banner:      input.Banner,
		ReleaseDate: releaseDate,
		InfluenceID: constants.GetInfluenceID(input.Influence),
		JobID:       constants.GetJobID(input.Job),
	}

	// Build accessory domain object if provided
	var updatedAccessory *domain.Accessory
	if input.Accessory != nil {
		updatedAccessory = &domain.Accessory{
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
	}

	// Update traveller with accessory in transaction
	// Repository handles checking if accessory exists and decides INSERT vs UPDATE
	err = s.travellerRepo.UpdateTravellerWithAccessory(ctx, id, &updatedTraveller, updatedAccessory)
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

func (s *travellerService) Delete(ctx context.Context, id int) (err error) {
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
