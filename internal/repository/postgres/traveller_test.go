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

	id := 1
	want := domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{ID: int64(id)}}

	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tr_traveller" WHERE id = $1 AND "tr_traveller"."deleted_at" IS NULL ORDER BY "tr_traveller"."id" LIMIT $2`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rarity"}).
			AddRow(1, want.Name, want.Rarity))

	res, err := s.repo.GetByID(context.TODO(), id)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), res, want)
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Create() {
	now := time.Now()

	traveller := &domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{
		CreatedAt: now,
		UpdatedAt: now,
	}}

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tr_traveller" ("created_by","updated_by","deleted_by","created_at","updated_at","deleted_at","name","rarity","influence_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
		WithArgs(traveller.CreatedBy, traveller.UpdatedBy, traveller.DeletedBy, traveller.CreatedAt, traveller.UpdatedAt, traveller.DeletedAt, traveller.Name, traveller.Rarity, traveller.InfluenceID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	s.mock.ExpectCommit()

	err := s.repo.Create(context.TODO(), traveller)
	assert.NoError(s.T(), err)
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Update() {

	traveller := &domain.Traveller{Name: "Fiore", Rarity: 5, CommonModel: domain.CommonModel{
		ID: int64(1),
	}}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tr_traveller" SET "updated_at"=$1,"name"=$2,"rarity"=$3 WHERE "tr_traveller"."deleted_at" IS NULL AND "id" = $4`)).WithArgs(helpers.AnyTime{}, traveller.Name, traveller.Rarity, traveller.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.mock.ExpectCommit()
	err := s.repo.Update(context.TODO(), traveller)
	assert.NoError(s.T(), err)
}

func (s *TravellerRepositorySuite) TestTravellerRepository_Delete() {

	id := 1
	s.mock.ExpectBegin()
	s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tr_traveller" SET "deleted_at"=$1 WHERE "tr_traveller"."id" = $2 AND "tr_traveller"."deleted_at" IS NULL`)).WithArgs(helpers.AnyTime{}, id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.mock.ExpectCommit()
	err := s.repo.Delete(context.TODO(), id)
	assert.NoError(s.T(), err)
}
