package accessory

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type AccessoryRepository interface {
	GetList(ctx context.Context, filter domain.ListAccessoryRequest, offset, limit int) (result []domain.Accessory, ownerNames map[int64]string, total int64, err error)
	Create(ctx context.Context, input *domain.Accessory) (err error)
	Update(ctx context.Context, input *domain.Accessory) (err error)
}

type Service struct {
	accessoryRepo AccessoryRepository
	logger        *logging.Logger
}

func NewAccessoryService(a AccessoryRepository, logger *logging.Logger) *Service {
	return &Service{
		accessoryRepo: a,
		logger:        logger.Named("service.accessory"),
	}
}

func (s Service) GetList(ctx context.Context, filter domain.ListAccessoryRequest, params helpers.PaginationParams) (res helpers.PaginatedResponse[domain.AccessoryListItemResponse], err error) {
	ctx, span := telemetry.StartServiceSpan(ctx, "service.accessory", "AccessoryService.GetList",
		attribute.Int("page", params.Page),
		attribute.Int("page_size", params.PageSize),
	)
	defer telemetry.EndSpanWithError(span, err)

	// Normalize pagination params
	params.Normalize()

	// Normalize order direction (case insensitive)
	if filter.OrderDir != "" {
		filter.OrderDir = strings.ToUpper(filter.OrderDir)
	}

	s.logger.WithContext(ctx).Info("fetching accessory list",
		zap.Int("page", params.Page),
		zap.Int("page_size", params.PageSize),
		zap.String("filter.owner", filter.Owner),
		zap.String("filter.effect", filter.Effect),
		zap.String("order_by", filter.OrderBy),
		zap.String("order_dir", filter.OrderDir),
	)

	accessories, ownerNames, total, err := s.accessoryRepo.GetList(ctx, filter, params.Offset(), params.PageSize)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to fetch accessory list",
			zap.String("error.type", "repository_error"),
			zap.String("error.message", err.Error()),
		)
		return
	}

	s.logger.WithContext(ctx).Info("accessory list fetched successfully",
		zap.Int64("total", total),
		zap.Int("returned", len(accessories)),
	)

	// Map to response DTOs
	items := make([]domain.AccessoryListItemResponse, len(accessories))
	for i, acc := range accessories {
		items[i] = domain.ToAccessoryListItemResponse(acc, ownerNames)
	}

	res = helpers.NewPaginatedResponse(items, params, total)

	return
}
