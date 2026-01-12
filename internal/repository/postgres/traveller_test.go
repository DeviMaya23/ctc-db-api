package postgres

import (
	"context"
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
	repo *TravellerRepository
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
		want    domain.Traveller
		wantErr bool
	}{
		{
			name: "found",
			id:   1,
			mockSet: func() {
				want := domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{ID: int64(1)}}
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_traveller" WHERE id = $1 AND "m_traveller"."deleted_at" IS NULL ORDER BY "m_traveller"."id" LIMIT $2`)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rarity"}).AddRow(1, want.Name, want.Rarity))
			},
			want:    domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{ID: int64(1)}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()

			res, err := s.repo.GetByID(context.TODO(), tt.id)
			if tt.wantErr {
				assert.Error(s.T(), err)
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

				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_traveller" WHERE "m_traveller"."deleted_at" IS NULL LIMIT $1`)).
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rarity"}).AddRow(1, "Fiore", 5).AddRow(2, "Shen", 4))
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

				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_traveller" WHERE LOWER(name) LIKE LOWER($1) AND influence_id = $2 AND job_id = $3 AND "m_traveller"."deleted_at" IS NULL LIMIT $4`)).
					WithArgs("%Fiore%", 1, 1, 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rarity", "job_id", "influence_id", "accessory_id"}).AddRow(1, "Fiore", 5, 1, 1, 0))

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
	tests := []struct {
		name      string
		traveller *domain.Traveller
		mockSet   func()
	}{
		{
			name:      "create success",
			traveller: &domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{CreatedAt: time.Now(), UpdatedAt: time.Now()}},
			mockSet: func() {
				t := &domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{CreatedAt: time.Now(), UpdatedAt: time.Now()}}
				s.mock.ExpectBegin()
				s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "m_traveller" ("created_by","updated_by","deleted_by","created_at","updated_at","deleted_at","name","rarity","influence_id","job_id","accessory_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
					WithArgs(t.CreatedBy, t.UpdatedBy, t.DeletedBy, t.CreatedAt, t.UpdatedAt, t.DeletedAt, t.Name, t.Rarity, t.InfluenceID, t.JobID, t.AccessoryID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				s.mock.ExpectCommit()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()
			err := s.repo.Create(context.TODO(), tt.traveller)
			assert.NoError(s.T(), err)
		})
	}
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Update() {
	tests := []struct {
		name      string
		traveller *domain.Traveller
		mockSet   func()
	}{
		{
			name:      "update success",
			traveller: &domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{ID: int64(1)}},
			mockSet: func() {
				t := &domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{ID: int64(1)}}
				s.mock.ExpectBegin()
				s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_traveller" SET "updated_at"=$1,"name"=$2,"rarity"=$3 WHERE "m_traveller"."deleted_at" IS NULL AND "id" = $4`)).WithArgs(helpers.AnyTime{}, t.Name, t.Rarity, t.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				s.mock.ExpectCommit()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()
			err := s.repo.Update(context.TODO(), tt.traveller)
			assert.NoError(s.T(), err)
		})
	}
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Delete() {
	tests := []struct {
		name    string
		id      int
		mockSet func()
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
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()
			err := s.repo.Delete(context.TODO(), tt.id)
			assert.NoError(s.T(), err)
		})
	}
}
