package traveller

import (
	"context"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/traveller/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TravellerServiceSuite struct {
	suite.Suite
	travellerRepo *mocks.MockTravellerRepository
	accessoryRepo *mocks.MockAccessoryRepository
	svc           *Service
}

func TestTravellerServiceSuite(t *testing.T) {
	suite.Run(t, new(TravellerServiceSuite))
}

func (s *TravellerServiceSuite) SetupSuite() {
	logger, _ := logging.NewDevelopmentLogger()

	s.travellerRepo = new(mocks.MockTravellerRepository)
	s.accessoryRepo = new(mocks.MockAccessoryRepository)
	s.svc = NewTravellerService(s.travellerRepo, s.accessoryRepo, logger)

}

func (s *TravellerServiceSuite) TestTravellerService_NewService() {

	s.T().Run("success", func(t *testing.T) {
		logger, _ := logging.NewDevelopmentLogger()
		repo := new(mocks.MockTravellerRepository)
		accessoryRepo := new(mocks.MockAccessoryRepository)
		NewTravellerService(repo, accessoryRepo, logger)
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
				Name:      "Viola",
				Rarity:    5,
				Influence: constants.InfluencePower,
				Job:       constants.JobWarrior,
			}},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Create", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "success with accessory",
			args: args{request: domain.CreateTravellerRequest{
				Name:      "Viola",
				Rarity:    5,
				Influence: constants.InfluencePower, Job: constants.JobWarrior, Accessory: &domain.CreateAccessoryRequest{
					Name: "Test Accessory",
				},
			}},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.accessoryRepo.On("Create", mock.Anything, mock.Anything).Return(want.err).Once()
				s.travellerRepo.On("Create", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "failed to create accessory",
			args: args{request: domain.CreateTravellerRequest{
				Name:      "Viola",
				Rarity:    5,
				Influence: constants.InfluencePower,
				Accessory: &domain.CreateAccessoryRequest{
					Name: "Test Accessory",
				},
			}},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.accessoryRepo.On("Create", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name:    "failed to create traveller",
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Create", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.svc.Create(ctx, tt.args.request)
			if tt.wantErr {
				assert.Equal(s.T(), err, tt.want.err)
				return
			}

			assert.Nil(s.T(), err)

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
					Name:      "Fiore Updated",
					Rarity:    5,
					Influence: constants.InfluencePower,
					Job:       constants.JobMerchant,
				},
			},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				// Mock GetByID to return existing traveller without accessory
				existingTraveller := domain.Traveller{
					CommonModel: domain.CommonModel{ID: int64(args.id)},
					Name:        "Fiore",
					Rarity:      5,
					InfluenceID: 1,
					AccessoryID: nil,
				}
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(existingTraveller, nil).Once()
				s.travellerRepo.On("Update", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "success with new accessory creation",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:      "Fiore Updated",
					Rarity:    5,
					Influence: constants.InfluencePower, Job: constants.JobMerchant, Accessory: &domain.UpdateAccessoryRequest{
						Name: "New Accessory",
						HP:   100,
					},
				},
			},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				// Mock GetByID to return existing traveller without accessory
				existingTraveller := domain.Traveller{
					CommonModel: domain.CommonModel{ID: int64(args.id)},
					Name:        "Fiore",
					Rarity:      5,
					InfluenceID: 1,
					AccessoryID: nil,
				}
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(existingTraveller, nil).Once()
				s.accessoryRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				s.travellerRepo.On("Update", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "success with existing accessory update",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:      "Fiore Updated",
					Rarity:    5,
					Influence: constants.InfluencePower, Job: constants.JobMerchant, Accessory: &domain.UpdateAccessoryRequest{
						Name: "Updated Accessory",
						HP:   200,
					},
				},
			},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				// Mock GetByID to return existing traveller with accessory
				accessoryID := 42
				existingTraveller := domain.Traveller{
					CommonModel: domain.CommonModel{ID: int64(args.id)},
					Name:        "Fiore",
					Rarity:      5,
					InfluenceID: 1,
					AccessoryID: &accessoryID,
				}
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(existingTraveller, nil).Once()
				s.accessoryRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				s.travellerRepo.On("Update", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "failed to get existing traveller",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:      "Fiore",
					Rarity:    5,
					Influence: constants.InfluencePower, Job: constants.JobMerchant},
			},
			want:    want{err: gorm.ErrRecordNotFound},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(domain.Traveller{}, want.err).Once()
			},
		}, {
			name: "failed to create accessory",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:      "Fiore",
					Rarity:    5,
					Influence: constants.InfluencePower, Job: constants.JobMerchant, Accessory: &domain.UpdateAccessoryRequest{
						Name: "New Accessory",
					},
				},
			},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				existingTraveller := domain.Traveller{
					CommonModel: domain.CommonModel{ID: int64(args.id)},
					Name:        "Fiore",
					AccessoryID: nil,
				}
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(existingTraveller, nil).Once()
				s.accessoryRepo.On("Create", mock.Anything, mock.Anything).Return(want.err).Once()
			},
		}, {
			name: "failed to update traveller",
			args: args{
				id: 1,
				input: domain.UpdateTravellerRequest{
					Name:      "Fiore",
					Rarity:    5,
					Influence: constants.InfluencePower, Job: constants.JobMerchant},
			},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				existingTraveller := domain.Traveller{
					CommonModel: domain.CommonModel{ID: int64(args.id)},
					Name:        "Fiore",
					AccessoryID: nil,
				}
				s.travellerRepo.On("GetByID", mock.Anything, args.id).Return(existingTraveller, nil).Once()
				s.travellerRepo.On("Update", mock.Anything, mock.Anything).Return(want.err).Once()
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
