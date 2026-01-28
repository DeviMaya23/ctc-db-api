package traveller

import (
	"context"
	"errors"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TravellerRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo *travellerRepository
}

func TestTravellerRepositorySuite(t *testing.T) {
	suite.Run(t, new(TravellerRepositorySuite))
}

func (s *TravellerRepositorySuite) SetupTest() {
	var err error
	s.db, s.mock, err = helpers.NewMockDB()
	if err != nil {
		s.T().Fatal()
	}

	logger, _ := logging.NewDevelopmentLogger()
	s.repo = NewTravellerRepository(s.db, logger)
}

func (s *TravellerRepositorySuite) TestTravellerRepository_GetByID() {
	tests := []struct {
		name    string
		id      int
		mockSet func()
		want    *domain.Traveller
		wantErr bool
		checkFn func(*testing.T, error)
	}{
		{
			name: "found",
			id:   1,
			mockSet: func() {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				want := domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(1)}}
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_traveller" WHERE id = $1 AND "m_traveller"."deleted_at" IS NULL ORDER BY "m_traveller"."id" LIMIT $2`)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rarity", "banner", "release_date"}).AddRow(1, want.Name, want.Rarity, want.Banner, want.ReleaseDate))
			},
			want: func() *domain.Traveller {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				return &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(1)}}
			}(),
			wantErr: false,
		},
		{
			name: "not found",
			id:   999,
			mockSet: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_traveller" WHERE id = $1 AND "m_traveller"."deleted_at" IS NULL ORDER BY "m_traveller"."id" LIMIT $2`)).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
			checkFn: func(t *testing.T, err error) {
				var nfe *domain.NotFoundError
				assert.True(t, errors.As(err, &nfe), "expected NotFoundError")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()

			res, err := s.repo.GetByID(context.TODO(), tt.id)
			if tt.wantErr {
				assert.Error(s.T(), err)
				if tt.checkFn != nil {
					tt.checkFn(s.T(), err)
				}
				return
			}
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.want, res)
		})
	}
}

func (s *TravellerRepositorySuite) TestTravellerRepository_GetList() {
	tests := []struct {
		name    string
		filter  domain.ListTravellerRequest
		offset  int
		limit   int
		mockSet func()
		wantTot int64
		wantLen int
	}{
		{
			name:   "no filters",
			filter: domain.ListTravellerRequest{},
			offset: 0,
			limit:  10,
			mockSet: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "m_traveller" WHERE "m_traveller"."deleted_at" IS NULL`)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				date1 := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				date2 := time.Date(2023, 6, 20, 0, 0, 0, 0, time.UTC)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_traveller" WHERE "m_traveller"."deleted_at" IS NULL LIMIT $1`)).
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rarity", "banner", "release_date"}).AddRow(1, "Fiore", 5, "General", date1).AddRow(2, "Shen", 4, "MT Orsterra", date2))
			},
			wantTot: 2,
			wantLen: 2,
		},
		{
			name: "with filters",
			filter: domain.ListTravellerRequest{
				Name:        "Fiore",
				JobID:       1,
				InfluenceID: 1,
			},
			offset: 0,
			limit:  10,
			mockSet: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "m_traveller" WHERE LOWER(name) LIKE LOWER($1) AND influence_id = $2 AND job_id = $3 AND "m_traveller"."deleted_at" IS NULL`)).
					WithArgs("%Fiore%", 1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_traveller" WHERE LOWER(name) LIKE LOWER($1) AND influence_id = $2 AND job_id = $3 AND "m_traveller"."deleted_at" IS NULL LIMIT $4`)).
					WithArgs("%Fiore%", 1, 1, 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rarity", "banner", "release_date", "job_id", "influence_id", "accessory_id"}).AddRow(1, "Fiore", 5, "General", releaseDate, 1, 1, 0))

				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_accessory" WHERE "m_accessory"."id" = $1 AND "m_accessory"."deleted_at" IS NULL`)).
					WithArgs(0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			wantTot: 1,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()

			result, total, err := s.repo.GetList(context.TODO(), tt.filter, tt.offset, tt.limit)
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.wantTot, total)
			assert.Equal(s.T(), tt.wantLen, len(result))

			if tt.wantLen > 0 {
				for i := 0; i < tt.wantLen; i++ {
					assert.Equal(s.T(), tt.filter.Name == "" || regexp.MustCompile("(?i)"+tt.filter.Name).MatchString(result[i].Name), true)
					if tt.filter.InfluenceID != 0 {
						assert.Equal(s.T(), result[i].InfluenceID, tt.filter.InfluenceID)
					}
					if tt.filter.JobID != 0 {
						assert.Equal(s.T(), result[i].JobID, tt.filter.JobID)
					}
				}
			}

		})
	}
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Create() {
	timeNow := time.Now()
	tests := []struct {
		name      string
		traveller *domain.Traveller
		mockSet   func()
		wantErr   bool
		checkFn   func(*testing.T, error)
	}{
		{
			name: "create success",
			traveller: func() *domain.Traveller {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				return &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{CreatedAt: timeNow, UpdatedAt: timeNow}}
			}(),
			mockSet: func() {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				t := &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{CreatedAt: timeNow, UpdatedAt: timeNow}}
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "m_traveller" ("created_by","updated_by","deleted_by","created_at","updated_at","deleted_at","name","rarity","banner","release_date","influence_id","job_id","accessory_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
					WithArgs(t.CreatedBy, t.UpdatedBy, t.DeletedBy, t.CreatedAt, t.UpdatedAt, t.DeletedAt, t.Name, t.Rarity, t.Banner, t.ReleaseDate, t.InfluenceID, t.JobID, t.AccessoryID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "duplicate name error",
			traveller: func() *domain.Traveller {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				return &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{CreatedAt: timeNow, UpdatedAt: timeNow}}
			}(),
			mockSet: func() {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				t := &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{CreatedAt: timeNow, UpdatedAt: timeNow}}
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "m_traveller" ("created_by","updated_by","deleted_by","created_at","updated_at","deleted_at","name","rarity","banner","release_date","influence_id","job_id","accessory_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
					WithArgs(t.CreatedBy, t.UpdatedBy, t.DeletedBy, t.CreatedAt, t.UpdatedAt, t.DeletedAt, t.Name, t.Rarity, t.Banner, t.ReleaseDate, t.InfluenceID, t.JobID, t.AccessoryID).
					WillReturnError(gorm.ErrDuplicatedKey)
				s.mock.ExpectRollback()
			},
			wantErr: true,
			checkFn: func(t *testing.T, err error) {
				var ce *domain.ConflictError
				assert.True(t, errors.As(err, &ce), "expected ConflictError")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()
			err := s.repo.Create(context.TODO(), tt.traveller)
			if tt.wantErr {
				assert.Error(s.T(), err)
				if tt.checkFn != nil {
					tt.checkFn(s.T(), err)
				}
				return
			}
			assert.NoError(s.T(), err)
		})
	}
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Update() {
	tests := []struct {
		name      string
		traveller *domain.Traveller
		mockSet   func()
		wantErr   bool
		checkFn   func(*testing.T, error)
	}{
		{
			name: "update success",
			traveller: func() *domain.Traveller {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				return &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(1)}}
			}(),
			mockSet: func() {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				t := &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(1)}}
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_traveller" SET "updated_at"=$1,"name"=$2,"rarity"=$3,"banner"=$4,"release_date"=$5 WHERE "m_traveller"."deleted_at" IS NULL AND "id" = $6`)).WithArgs(helpers.AnyTime{}, t.Name, t.Rarity, t.Banner, t.ReleaseDate, t.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "not found",
			traveller: func() *domain.Traveller {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				return &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(999)}}
			}(),
			mockSet: func() {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				t := &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(999)}}
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_traveller" SET "updated_at"=$1,"name"=$2,"rarity"=$3,"banner"=$4,"release_date"=$5 WHERE "m_traveller"."deleted_at" IS NULL AND "id" = $6`)).WithArgs(helpers.AnyTime{}, t.Name, t.Rarity, t.Banner, t.ReleaseDate, t.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				s.mock.ExpectCommit()
			},
			wantErr: true,
			checkFn: func(t *testing.T, err error) {
				var nfe *domain.NotFoundError
				assert.True(t, errors.As(err, &nfe), "expected NotFoundError")
			},
		},
		{
			name: "duplicate name error",
			traveller: func() *domain.Traveller {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				return &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(1)}}
			}(),
			mockSet: func() {
				releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
				t := &domain.Traveller{Name: "Fiore", Rarity: 5, Banner: "General", ReleaseDate: releaseDate, CommonModel: domain.CommonModel{ID: int64(1)}}
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_traveller" SET "updated_at"=$1,"name"=$2,"rarity"=$3,"banner"=$4,"release_date"=$5 WHERE "m_traveller"."deleted_at" IS NULL AND "id" = $6`)).WithArgs(helpers.AnyTime{}, t.Name, t.Rarity, t.Banner, t.ReleaseDate, t.ID).
					WillReturnError(gorm.ErrDuplicatedKey)
				s.mock.ExpectRollback()
			},
			wantErr: true,
			checkFn: func(t *testing.T, err error) {
				var ce *domain.ConflictError
				assert.True(t, errors.As(err, &ce), "expected ConflictError")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()
			err := s.repo.Update(context.TODO(), tt.traveller)
			if tt.wantErr {
				assert.Error(s.T(), err)
				if tt.checkFn != nil {
					tt.checkFn(s.T(), err)
				}
				return
			}
			assert.NoError(s.T(), err)
		})
	}
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Delete() {
	tests := []struct {
		name    string
		id      int
		mockSet func()
		wantErr bool
		checkFn func(*testing.T, error)
	}{
		{
			name: "delete success",
			id:   1,
			mockSet: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_traveller" SET "deleted_at"=$1 WHERE "m_traveller"."id" = $2 AND "m_traveller"."deleted_at" IS NULL`)).WithArgs(helpers.AnyTime{}, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				s.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   999,
			mockSet: func() {
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_traveller" SET "deleted_at"=$1 WHERE "m_traveller"."id" = $2 AND "m_traveller"."deleted_at" IS NULL`)).WithArgs(helpers.AnyTime{}, 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
				s.mock.ExpectCommit()
			},
			wantErr: true,
			checkFn: func(t *testing.T, err error) {
				var nfe *domain.NotFoundError
				assert.True(t, errors.As(err, &nfe), "expected NotFoundError")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()
			err := s.repo.Delete(context.TODO(), tt.id)
			if tt.wantErr {
				assert.Error(s.T(), err)
				if tt.checkFn != nil {
					tt.checkFn(s.T(), err)
				}
				return
			}
			assert.NoError(s.T(), err)
		})
	}
}
