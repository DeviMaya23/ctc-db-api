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

type AccessoryRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo *AccessoryRepository
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
