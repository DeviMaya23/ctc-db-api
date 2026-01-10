package postgres

import (
	"context"
	"database/sql"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"

	"github.com/stretchr/testify/assert"

	pgGormDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUserRepository_Integration(t *testing.T) {
	ctx := context.Background()

	connStr := helpers.GetTestDB(t)

	dbConn, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatal("failed open database ", err)
	}
	db, err := gorm.Open(pgGormDriver.New(pgGormDriver.Config{
		Conn: dbConn,
	}), &gorm.Config{})
	if err != nil {
		t.Fatal("failed to open gorm ", err)
	}

	logger, _ := logging.NewDevelopmentLogger()
	repo := NewUserRepository(db, logger)

	// existing user
	user, err := repo.GetByUsername(ctx, "isla")
	assert.Nil(t, err)
	assert.Equal(t, user.Username, "isla")

	// not found
	user, err = repo.GetByUsername(ctx, "klins")
	assert.Equal(t, err, gorm.ErrRecordNotFound)

}
