package traveller

import (
	"context"
	"lizobly/ctc-db-api/internal/traveller/mocks"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TravellerServiceSuite struct {
	suite.Suite
	travellerRepo *mocks.MockTravellerRepository
	svc           *travellerService
}

func TestTravellerServiceSuite(t *testing.T) {
	suite.Run(t, new(TravellerServiceSuite))
}

func (s *TravellerServiceSuite) SetupTest() {
	logger, _ := logging.NewDevelopmentLogger()

	s.travellerRepo = new(mocks.MockTravellerRepository)
	s.svc = NewTravellerService(s.travellerRepo, logger)
}

func (s *TravellerServiceSuite) TearDownTest() {
	s.travellerRepo.AssertExpectations(s.T())
}

func (s *TravellerServiceSuite) TestTravellerService_NewService() {

	s.T().Run("success", func(t *testing.T) {
		logger, _ := logging.NewDevelopmentLogger()
		repo := new(mocks.MockTravellerRepository)
		NewTravellerService(repo, logger)
	})
}

func (s *TravellerServiceSuite) TestTravellerService_GetByID() {
	type args struct {
		id int
	}
	type want struct {
		traveller domain.Traveller
		err       error
	}
	tests := []struct {
		name       string
		args       args
		want       want
		wantErr    bool
		beforeTest func(ctx context.Context, args args, want want)
	}{
		{
			name: "success",
			args: args{id: 1},
			want: want{traveller: domain.Traveller{
				Name: "Fiore",
				CommonModel: domain.CommonModel{
					ID: 1,
				},
			}},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(want.traveller, want.err).Once()

			},
		}, {
			name:    "failed",
			args:    args{id: 1},
			want:    want{err: gorm.ErrRecordNotFound},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(want.traveller, want.err).Once()

			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			got, err := s.svc.GetByID(ctx, tt.args.id)
			if tt.wantErr {
				assert.Equal(s.T(), err, tt.want.err)
				return
			}

			assert.Nil(s.T(), err)
			assert.Equal(s.T(), got, tt.want.traveller)

		})
	}
}

func (s *TravellerServiceSuite) TestTravellerService_Create() {
	type args struct {
		request domain.CreateTravellerRequest
	}
	type want struct {
		err error
	}
	tests := []struct {
		name       string
		args       args
		want       want
		wantErr    bool
		beforeTest func(ctx context.Context, args args, want want)
	}{
		{
			name: "success without accessory",
			args: args{request: domain.CreateTravellerRequest{
				Name:        "Viola",
				Rarity:      5,
				Banner:      "General",
				ReleaseDate: "15-05-2023",
				Influence:   constants.InfluencePower,
				Job:         constants.JobWarrior,
			}},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("CreateTravellerWithAccessory", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					traveller := args.Get(1).(*domain.Traveller)
					traveller.ID = 123
				}).Return(want.err).Once()
			},
		}, {
			name: "success with accessory",
			args: args{request: domain.CreateTravellerRequest{
				Name:        "Viola",
				Rarity:      5,
				Banner:      "General",
				ReleaseDate: "15-05-2023",
				Influence:   constants.InfluencePower,
				Job:         constants.JobWarrior,
				Accessory: &domain.CreateAccessoryRequest{
					Name: "Test Accessory",
				},
			}},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("CreateTravellerWithAccessory", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					traveller := args.Get(1).(*domain.Traveller)
					accessory := args.Get(2).(*domain.Accessory)
					traveller.ID = 123
					accessory.ID = 456
				}).Return(want.err).Once()
			},
		}, {
			name: "failed to create accessory",
			args: args{request: domain.CreateTravellerRequest{
				Name:        "Viola",
				Rarity:      5,
				Banner:      "General",
				ReleaseDate: "15-05-2023",
				Influence:   constants.InfluencePower,
				Job:         constants.JobWarrior,
				Accessory: &domain.CreateAccessoryRequest{
					Name: "Test Accessory",
				},
			}},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("CreateTravellerWithAccessory", mock.Anything, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "failed to create traveller",
			args: args{request: domain.CreateTravellerRequest{
				Name:        "Viola",
				Rarity:      5,
				Banner:      "General",
				ReleaseDate: "15-05-2023",
				Influence:   constants.InfluencePower,
				Job:         constants.JobWarrior,
			}},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("CreateTravellerWithAccessory", mock.Anything, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			id, err := s.svc.Create(ctx, tt.args.request)
			if tt.wantErr {
				assert.Equal(s.T(), err, tt.want.err)
				return
			}

			assert.Nil(s.T(), err)
			assert.Greater(s.T(), id, int64(0))

		})
	}
}

func (s *TravellerServiceSuite) TestTravellerService_Update() {
	type args struct {
		id    int
		input domain.UpdateTravellerRequest
	}
	type want struct {
		err error
	}
	tests := []struct {
		name       string
		args       args
		want       want
		wantErr    bool
		beforeTest func(ctx context.Context, args args, want want)
	}{
		{
			name: "success without accessory",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:        "Fiore Updated",
					Rarity:      5,
					Banner:      "General",
					ReleaseDate: "15-05-2023",
					Influence:   constants.InfluencePower,
					Job:         constants.JobMerchant,
				},
			},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("UpdateTravellerWithAccessory", mock.Anything, args.id, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "success with new accessory creation",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:        "Fiore Updated",
					Rarity:      5,
					Banner:      "General",
					ReleaseDate: "15-05-2023",
					Influence:   constants.InfluencePower,
					Job:         constants.JobMerchant,
					Accessory: &domain.UpdateAccessoryRequest{
						Name: "New Accessory",
						HP:   100,
					},
				},
			},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("UpdateTravellerWithAccessory", mock.Anything, args.id, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "success with existing accessory update",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:        "Fiore Updated",
					Rarity:      5,
					Banner:      "General",
					ReleaseDate: "15-05-2023",
					Influence:   constants.InfluencePower,
					Job:         constants.JobMerchant,
					Accessory: &domain.UpdateAccessoryRequest{
						Name: "Updated Accessory",
						HP:   200,
					},
				},
			},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("UpdateTravellerWithAccessory", mock.Anything, args.id, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "failed to get existing traveller",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:        "Fiore",
					Rarity:      5,
					Banner:      "General",
					ReleaseDate: "15-05-2023",
					Influence:   constants.InfluencePower,
					Job:         constants.JobMerchant,
				},
			},
			want:    want{err: gorm.ErrRecordNotFound},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("UpdateTravellerWithAccessory", mock.Anything, args.id, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "failed to create accessory",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:        "Fiore",
					Rarity:      5,
					Banner:      "General",
					ReleaseDate: "15-05-2023",
					Influence:   constants.InfluencePower,
					Job:         constants.JobMerchant,
					Accessory: &domain.UpdateAccessoryRequest{
						Name: "New Accessory",
					},
				},
			},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("UpdateTravellerWithAccessory", mock.Anything, args.id, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "failed to update traveller",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:        "Fiore",
					Rarity:      5,
					Banner:      "General",
					ReleaseDate: "15-05-2023",
					Influence:   constants.InfluencePower,
					Job:         constants.JobMerchant,
				},
			},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("UpdateTravellerWithAccessory", mock.Anything, args.id, mock.Anything, mock.Anything).Return(want.err).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.svc.Update(ctx, tt.args.id, tt.args.input)
			if tt.wantErr {
				assert.Equal(s.T(), err, tt.want.err)
				return
			}

			assert.Nil(s.T(), err)

		})
	}
}

func (s *TravellerServiceSuite) TestTravellerService_Delete() {
	type args struct {
		request int
	}
	type want struct {
		err error
	}
	tests := []struct {
		name       string
		args       args
		want       want
		wantErr    bool
		beforeTest func(ctx context.Context, args args, want want)
	}{
		{
			name:    "success",
			args:    args{request: 1},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Delete", mock.Anything, args.request).Return(want.err).Once()

			},
		}, {
			name:    "failed",
			args:    args{request: 2},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Delete", mock.Anything, args.request).Return(want.err).Once()

			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.svc.Delete(ctx, tt.args.request)
			if tt.wantErr {
				assert.Equal(s.T(), err, tt.want.err)
				return
			}

			assert.Nil(s.T(), err)

		})
	}
}

func (s *TravellerServiceSuite) TestTravellerService_GetList() {
	type args struct {
		filter domain.ListTravellerRequest
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
				filter: domain.ListTravellerRequest{},
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
				travellers := []domain.Traveller{
					{CommonModel: domain.CommonModel{ID: 1}, Name: "Fiore", Rarity: 5},
					{CommonModel: domain.CommonModel{ID: 2}, Name: "Viola", Rarity: 4},
				}
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(travellers, want.total, want.err).Once()
			},
		},
		{
			name: "success with name filter",
			args: args{
				filter: domain.ListTravellerRequest{Name: "Fiore"},
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
				travellers := []domain.Traveller{
					{CommonModel: domain.CommonModel{ID: 1}, Name: "Fiore", Rarity: 5},
				}
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(travellers, want.total, want.err).Once()
			},
		},
		{
			name: "success with influence filter",
			args: args{
				filter: domain.ListTravellerRequest{Influence: constants.InfluencePower, InfluenceID: constants.GetInfluenceID(constants.InfluencePower)},
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
				travellers := []domain.Traveller{
					{CommonModel: domain.CommonModel{ID: 1}, Name: "Fiore", Rarity: 5, InfluenceID: constants.GetInfluenceID(constants.InfluencePower)},
				}
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(travellers, want.total, want.err).Once()
			},
		},
		{
			name: "success with job filter",
			args: args{
				filter: domain.ListTravellerRequest{Job: constants.JobWarrior, JobID: constants.GetJobID(constants.JobWarrior)},
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
				travellers := []domain.Traveller{
					{CommonModel: domain.CommonModel{ID: 1}, Name: "Fiore", Rarity: 5, JobID: constants.GetJobID(constants.JobWarrior)},
				}
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(travellers, want.total, want.err).Once()
			},
		},
		{
			name: "success with pagination defaults applied",
			args: args{
				filter: domain.ListTravellerRequest{},
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
				travellers := make([]domain.Traveller, 10)
				for i := 0; i < 10; i++ {
					travellers[i] = domain.Traveller{CommonModel: domain.CommonModel{ID: int64(i + 1)}, Name: "Test"}
				}
				// Normalized params: page 1, page_size 10, offset 0
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(travellers, want.total, want.err).Once()
			},
		},
		{
			name: "success with pagination page 2",
			args: args{
				filter: domain.ListTravellerRequest{},
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
				travellers := make([]domain.Traveller, 10)
				for i := 0; i < 10; i++ {
					travellers[i] = domain.Traveller{CommonModel: domain.CommonModel{ID: int64(i + 11)}, Name: "Test"}
				}
				// Page 2: offset = (2-1)*10 = 10
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 10, 10).Return(travellers, want.total, want.err).Once()
			},
		},
		{
			name: "success with empty result",
			args: args{
				filter: domain.ListTravellerRequest{Name: "NonExistent"},
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
				travellers := []domain.Traveller{}
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(travellers, want.total, want.err).Once()
			},
		},
		{
			name: "failed to fetch list",
			args: args{
				filter: domain.ListTravellerRequest{},
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
				s.travellerRepo.On("GetList", mock.Anything, args.filter, 0, 10).Return(nil, int64(0), want.err).Once()
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
				assert.Greater(s.T(), result.TotalPages, -1)
			}
		})
	}
}
