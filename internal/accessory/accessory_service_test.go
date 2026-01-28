package accessory

import (
	"context"
	"lizobly/ctc-db-api/internal/accessory/mocks"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AccessoryServiceSuite struct {
	suite.Suite
	accessoryRepo *mocks.MockAccessoryRepository
	svc           *accessoryService
}

func TestAccessoryServiceSuite(t *testing.T) {
	suite.Run(t, new(AccessoryServiceSuite))
}

func (s *AccessoryServiceSuite) SetupTest() {
	logger, _ := logging.NewDevelopmentLogger()

	s.accessoryRepo = new(mocks.MockAccessoryRepository)
	s.svc = NewAccessoryService(s.accessoryRepo, logger)
}

func (s *AccessoryServiceSuite) TearDownTest() {
	s.accessoryRepo.AssertExpectations(s.T())
}

func (s *AccessoryServiceSuite) TestAccessoryService_NewService() {
	s.T().Run("success", func(t *testing.T) {
		logger, _ := logging.NewDevelopmentLogger()
		repo := new(mocks.MockAccessoryRepository)
		svc := NewAccessoryService(repo, logger)
		assert.NotNil(t, svc)
	})
}

func (s *AccessoryServiceSuite) TestAccessoryService_GetList() {
	type args struct {
		filter domain.ListAccessoryRequest
		params helpers.PaginationParams
	}
	type want struct {
		count         int
		total         int64
		err           error
		hasPagination bool
	}
	tests := []struct {
		name       string
		args       args
		want       want
		wantErr    bool
		beforeTest func(ctx context.Context, args args, want want)
	}{
		{
			name: "success with no filters",
			args: args{
				filter: domain.ListAccessoryRequest{},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         2,
				total:         2,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := map[int64]string{
					1: "Fiore",
					2: "Viola",
				}
				accessories := []domain.Accessory{
					{
						CommonModel: domain.CommonModel{ID: 1},
						Name:        "Sword",
						Effect:      "ATK+10",
					},
					{
						CommonModel: domain.CommonModel{ID: 2},
						Name:        "Shield",
						Effect:      "DEF+10",
					},
				}
				s.accessoryRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "success with owner filter",
			args: args{
				filter: domain.ListAccessoryRequest{Owner: "Fiore"},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         1,
				total:         1,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := map[int64]string{
					1: "Fiore",
				}
				accessories := []domain.Accessory{
					{
						CommonModel: domain.CommonModel{ID: 1},
						Name:        "Sword",
						Effect:      "ATK+10",
					},
				}
				s.accessoryRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "success with effect filter",
			args: args{
				filter: domain.ListAccessoryRequest{Effect: "ATK"},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         1,
				total:         1,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := map[int64]string{
					1: "Fiore",
				}
				accessories := []domain.Accessory{
					{
						CommonModel: domain.CommonModel{ID: 1},
						Name:        "Sword",
						Effect:      "ATK+10",
					},
				}
				s.accessoryRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "success with ordering asc",
			args: args{
				filter: domain.ListAccessoryRequest{
					OrderBy:  "hp",
					OrderDir: "asc",
				},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         2,
				total:         2,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := map[int64]string{
					1: "Fiore",
					2: "Viola",
				}
				accessories := []domain.Accessory{
					{
						CommonModel: domain.CommonModel{ID: 1},
						Name:        "Sword",
						HP:          10,
						Effect:      "ATK+10",
					},
					{
						CommonModel: domain.CommonModel{ID: 2},
						Name:        "Shield",
						HP:          20,
						Effect:      "DEF+10",
					},
				}
				// Note: service normalizes order direction to uppercase
				normalizedFilter := args.filter
				normalizedFilter.OrderDir = "ASC"
				s.accessoryRepo.On("GetList", mock.Anything, normalizedFilter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "success with ordering desc",
			args: args{
				filter: domain.ListAccessoryRequest{
					OrderBy:  "sp",
					OrderDir: "desc",
				},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         2,
				total:         2,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := map[int64]string{
					1: "Fiore",
					2: "Viola",
				}
				accessories := []domain.Accessory{
					{
						CommonModel: domain.CommonModel{ID: 2},
						Name:        "Shield",
						SP:          30,
						Effect:      "DEF+10",
					},
					{
						CommonModel: domain.CommonModel{ID: 1},
						Name:        "Sword",
						SP:          20,
						Effect:      "ATK+10",
					},
				}
				// Note: service normalizes order direction to uppercase
				normalizedFilter := args.filter
				normalizedFilter.OrderDir = "DESC"
				s.accessoryRepo.On("GetList", mock.Anything, normalizedFilter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "success with pagination defaults applied",
			args: args{
				filter: domain.ListAccessoryRequest{},
				params: helpers.PaginationParams{Page: 0, PageSize: 0}, // Invalid params should be normalized
			},
			want: want{
				count:         10,
				total:         50,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := make(map[int64]string)
				accessories := make([]domain.Accessory, 10)
				for i := 0; i < 10; i++ {
					accessories[i] = domain.Accessory{
						CommonModel: domain.CommonModel{ID: int64(i + 1)},
						Name:        "Accessory",
						Effect:      "Effect",
					}
					ownerNames[int64(i+1)] = "Owner"
				}
				// Normalized params: page 1, page_size 10, offset 0
				s.accessoryRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "success with pagination page 2",
			args: args{
				filter: domain.ListAccessoryRequest{},
				params: helpers.PaginationParams{Page: 2, PageSize: 10},
			},
			want: want{
				count:         10,
				total:         25,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := make(map[int64]string)
				accessories := make([]domain.Accessory, 10)
				for i := 0; i < 10; i++ {
					accessories[i] = domain.Accessory{
						CommonModel: domain.CommonModel{ID: int64(i + 11)},
						Name:        "Accessory",
						Effect:      "Effect",
					}
					ownerNames[int64(i+11)] = "Owner"
				}
				// Page 2: offset = (2-1)*10 = 10
				s.accessoryRepo.On("GetList", mock.Anything, args.filter, 10, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "success with empty result",
			args: args{
				filter: domain.ListAccessoryRequest{Effect: "NonExistent"},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         0,
				total:         0,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				accessories := []domain.Accessory{}
				ownerNames := map[int64]string{}
				s.accessoryRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
		{
			name: "failed to fetch list",
			args: args{
				filter: domain.ListAccessoryRequest{},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         0,
				total:         0,
				err:           gorm.ErrInvalidDB,
				hasPagination: false,
			},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.accessoryRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(nil, nil, int64(0), want.err).Once()
			},
		},
		{
			name: "success with all filters combined",
			args: args{
				filter: domain.ListAccessoryRequest{
					Owner:    "Fiore",
					Effect:   "ATK",
					OrderBy:  "patk",
					OrderDir: "asc",
				},
				params: helpers.PaginationParams{Page: 1, PageSize: 10},
			},
			want: want{
				count:         1,
				total:         1,
				err:           nil,
				hasPagination: true,
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				ownerNames := map[int64]string{
					1: "Fiore",
				}
				accessories := []domain.Accessory{
					{
						CommonModel: domain.CommonModel{ID: 1},
						Name:        "Sword",
						PAtk:        50,
						Effect:      "ATK+10",
					},
				}
				normalizedFilter := args.filter
				normalizedFilter.OrderDir = "ASC"
				s.accessoryRepo.On("GetList", mock.Anything, normalizedFilter, 0, 10).Return(accessories, ownerNames, want.total, want.err).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			result, err := s.svc.GetList(ctx, tt.args.filter, tt.args.params)
			if tt.wantErr {
				assert.Equal(s.T(), err, tt.want.err)
				return
			}

			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.want.count, len(result.Data))
			assert.Equal(s.T(), tt.want.total, result.Total)
			if tt.want.hasPagination {
				assert.Greater(s.T(), result.Page, 0)
				assert.Greater(s.T(), result.PageSize, 0)
				assert.GreaterOrEqual(s.T(), result.TotalPages, 0)
			}

			// Verify owner names are mapped correctly
			for i, item := range result.Data {
				if i < len(result.Data) && item.Owner != "" {
					assert.NotEmpty(s.T(), item.Owner)
				}
			}
		})
	}
}
