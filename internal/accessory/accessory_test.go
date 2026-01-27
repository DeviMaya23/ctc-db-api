package accessory

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

type AccessoryRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo *accessoryRepository
}

func TestAccessoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(AccessoryRepositorySuite))
}

func (s *AccessoryRepositorySuite) SetupTest() {
	var err error
	s.db, s.mock, err = helpers.NewMockDB()
	if err != nil {
		s.T().Fatal()
	}

	logger, _ := logging.NewDevelopmentLogger()
	s.repo = NewAccessoryRepository(s.db, logger)
}

func (s *AccessoryRepositorySuite) TestAccessoryRepository_Create() {
	now := time.Now()

	accessory := &domain.Accessory{
		Name:   "Crown of Wisdom",
		HP:     150,
		SP:     80,
		PAtk:   45,
		PDef:   30,
		EAtk:   60,
		EDef:   25,
		Spd:    12,
		Crit:   8,
		Effect: "Increases elemental damage by 15%",
		CommonModel: domain.CommonModel{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "m_accessory" ("created_by","updated_by","deleted_by","created_at","updated_at","deleted_at","name","hp","sp","patk","pdef","eatk","edef","spd","crit","effect") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING "id"`)).
		WithArgs(accessory.CreatedBy, accessory.UpdatedBy, accessory.DeletedBy, accessory.CreatedAt, accessory.UpdatedAt, accessory.DeletedAt, accessory.Name, accessory.HP, accessory.SP, accessory.PAtk, accessory.PDef, accessory.EAtk, accessory.EDef, accessory.Spd, accessory.Crit, accessory.Effect).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	s.mock.ExpectCommit()

	err := s.repo.Create(context.TODO(), accessory)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), accessory.ID)
}

func (s *AccessoryRepositorySuite) TestAccessoryRepository_GetList() {
	tests := []struct {
		name    string
		filter  domain.ListAccessoryRequest
		offset  int
		limit   int
		mockSet func()
		wantTot int64
		wantLen int
	}{
		{
			name:   "no filters",
			filter: domain.ListAccessoryRequest{},
			offset: 0,
			limit:  10,
			mockSet: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "m_accessory" LEFT JOIN m_traveller ON m_accessory.id = m_traveller.accessory_id WHERE "m_accessory"."deleted_at" IS NULL`)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT m_accessory.*, m_traveller.name as owner FROM "m_accessory" LEFT JOIN m_traveller ON m_accessory.id = m_traveller.accessory_id WHERE "m_accessory"."deleted_at" IS NULL LIMIT $1`)).
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "hp", "sp", "patk", "pdef", "eatk", "edef", "spd", "crit", "effect", "owner"}).
						AddRow(1, "Crown of Wisdom", 150, 80, 45, 30, 60, 25, 12, 8, "Increases elemental damage", "Fiore").
						AddRow(2, "Ring of Power", 200, 100, 80, 50, 70, 40, 15, 10, "Increases physical damage", "Noctis"))
			},
			wantTot: 2,
			wantLen: 2,
		},
		{
			name: "with owner filter",
			filter: domain.ListAccessoryRequest{
				Owner: "Fiore",
			},
			offset: 0,
			limit:  10,
			mockSet: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "m_accessory" LEFT JOIN m_traveller ON m_accessory.id = m_traveller.accessory_id WHERE LOWER(m_traveller.name) LIKE LOWER($1) AND "m_accessory"."deleted_at" IS NULL`)).
					WithArgs("%Fiore%").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT m_accessory.*, m_traveller.name as owner FROM "m_accessory" LEFT JOIN m_traveller ON m_accessory.id = m_traveller.accessory_id WHERE LOWER(m_traveller.name) LIKE LOWER($1) AND "m_accessory"."deleted_at" IS NULL LIMIT $2`)).
					WithArgs("%Fiore%", 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "hp", "sp", "patk", "pdef", "eatk", "edef", "spd", "crit", "effect", "owner"}).
						AddRow(1, "Crown of Wisdom", 150, 80, 45, 30, 60, 25, 12, 8, "Increases elemental damage", "Fiore"))
			},
			wantTot: 1,
			wantLen: 1,
		},
		{
			name: "with effect and ordering filters",
			filter: domain.ListAccessoryRequest{
				Effect:   "Elemental",
				OrderBy:  "hp",
				OrderDir: "DESC",
			},
			offset: 0,
			limit:  10,
			mockSet: func() {
				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "m_accessory" LEFT JOIN m_traveller ON m_accessory.id = m_traveller.accessory_id WHERE LOWER(m_accessory.effect) LIKE LOWER($1) AND "m_accessory"."deleted_at" IS NULL`)).
					WithArgs("%Elemental%").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT m_accessory.*, m_traveller.name as owner FROM "m_accessory" LEFT JOIN m_traveller ON m_accessory.id = m_traveller.accessory_id WHERE LOWER(m_accessory.effect) LIKE LOWER($1) AND "m_accessory"."deleted_at" IS NULL ORDER BY m_accessory.hp DESC LIMIT $2`)).
					WithArgs("%Elemental%", 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "hp", "sp", "patk", "pdef", "eatk", "edef", "spd", "crit", "effect", "owner"}).
						AddRow(1, "Crown of Wisdom", 150, 80, 45, 30, 60, 25, 12, 8, "Increases elemental damage by 15%", "Fiore"))
			},
			wantTot: 1,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.mockSet()

			result, ownerNames, total, err := s.repo.GetList(context.TODO(), tt.filter, tt.offset, tt.limit)
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.wantTot, total)
			assert.Equal(s.T(), tt.wantLen, len(result))

			if tt.wantLen > 0 {
				for i := 0; i < tt.wantLen; i++ {
					if tt.filter.Owner != "" {
						assert.Equal(s.T(), regexp.MustCompile("(?i)"+tt.filter.Owner).MatchString(ownerNames[result[i].ID]), true)
					}
					if tt.filter.Effect != "" {
						assert.Equal(s.T(), regexp.MustCompile("(?i)"+tt.filter.Effect).MatchString(result[i].Effect), true)
					}
				}
			}
		})
	}
}
