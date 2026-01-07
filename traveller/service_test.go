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
	svc           *Service
}

func TestTravellerServiceSuite(t *testing.T) {
	suite.Run(t, new(TravellerServiceSuite))
}

func (s *TravellerServiceSuite) SetupSuite() {
	logger, _ := logging.NewDevelopmentLogger()

	s.travellerRepo = new(mocks.MockTravellerRepository)
	s.svc = NewTravellerService(s.travellerRepo, logger)

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
				s.travellerRepo.On("GetByID", ctx, args.id).Return(want.traveller, want.err).Once()

			},
		}, {
			name:    "failed",
			args:    args{id: 1},
			want:    want{err: gorm.ErrRecordNotFound},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("GetByID", ctx, args.id).Return(want.traveller, want.err).Once()

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
			name: "success",
			args: args{request: domain.CreateTravellerRequest{
				Name:      "Viola",
				Rarity:    5,
				Influence: constants.InfluencePower,
			}},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				newTraveller := &domain.Traveller{
					Name:        args.request.Name,
					Rarity:      args.request.Rarity,
					InfluenceID: constants.GetInfluenceID(args.request.Influence),
				}
				s.travellerRepo.On("Create", ctx, newTraveller).Return(want.err).Once()

			},
		}, {
			name:    "failed",
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Create", ctx, mock.Anything).Return(want.err).Once()

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
		request *domain.Traveller
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
			args:    args{request: &domain.Traveller{Name: "Fiore"}},
			want:    want{},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Update", ctx, args.request).Return(want.err).Once()

			},
		}, {
			name:    "failed",
			args:    args{request: &domain.Traveller{Name: "Fiore", CommonModel: domain.CommonModel{ID: 1}}},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Update", ctx, args.request).Return(want.err).Once()

			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.svc.Update(ctx, tt.args.request)
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
				s.travellerRepo.On("Delete", ctx, args.request).Return(want.err).Once()

			},
		}, {
			name:    "failed",
			args:    args{request: 2},
			want:    want{err: gorm.ErrInvalidDB},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.travellerRepo.On("Delete", ctx, args.request).Return(want.err).Once()

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
