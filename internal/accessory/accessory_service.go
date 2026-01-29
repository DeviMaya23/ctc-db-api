package accessory

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type AccessoryRepository interface {
	GetList(ctx context.Context, filter domain.ListAccessoryRequest, offset, limit int) (result []*domain.Accessory, ownerNames map[int64]string, total int64, err error)
	Create(ctx context.Context, input *domain.Accessory) (err error)
	Update(ctx context.Context, input *domain.Accessory) (err error)
}

type accessoryService struct {
	accessoryRepo AccessoryRepository
	logger        *logging.Logger
}

func NewAccessoryService(a AccessoryRepository, logger *logging.Logger) *accessoryService {
	return &accessoryService{
		accessoryRepo: a,
		logger:        logger.Named("service.accessory"),
	}
}

func (s *accessoryService) GetList(ctx context.Context, filter domain.ListAccessoryRequest, params helpers.PaginationParams) (res helpers.PaginatedResponse[domain.AccessoryListItemResponse], err error) {
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

	accessories, ownerNames, total, err := s.accessoryRepo.GetList(ctx, filter, params.Offset(), params.PageSize)
	if err != nil {
		return
	}

	// Map to response DTOs
	items := make([]domain.AccessoryListItemResponse, len(accessories))
	for i, acc := range accessories {
		items[i] = domain.ToAccessoryListItemResponse(acc, ownerNames)
	}

	res = helpers.NewPaginatedResponse(items, params, total)

	return
}
