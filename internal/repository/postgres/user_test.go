package postgres

import (
	"context"
	"lizobly/cotc-db-api/pkg/domain"
	"lizobly/cotc-db-api/pkg/helpers"
	"lizobly/cotc-db-api/pkg/logging"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
	repo *UserRepository
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

func (s *UserRepositorySuite) SetupTest() {
	var err error
	s.db, s.mock, err = helpers.NewMockDB()
	if err != nil {
		s.T().Fatal()
	}

	logger, _ := logging.NewDevelopmentLogger()
	s.repo = NewUserRepository(s.db, logger)
}

func (s *UserRepositorySuite) TestUserRepository_GetByUsername() {

	username := "uname"
	want := domain.User{Username: username}

	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "m_user" WHERE username = $1 AND "m_user"."deleted_at" IS NULL ORDER BY "m_user"."id" LIMIT $2`)).
		WillReturnRows(sqlmock.NewRows([]string{"username"}).
			AddRow(want.Username))

	res, err := s.repo.GetByUsername(context.TODO(), username)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), res, want)
}
