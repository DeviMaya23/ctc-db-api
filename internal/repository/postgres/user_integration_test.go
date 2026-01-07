package postgres

import (
	"context"
	"database/sql"
	"lizobly/ctc-db-api/pkg/logging"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	pgGormDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUserRepository_Integration(t *testing.T) {
	ctx := context.Background()

	initDBFilePath := filepath.Join("../../..", "testdata", "db-user-repo.sql")

	pgContainer, err := postgres.Run(ctx, "postgres:15.3-alpine",
		postgres.WithInitScripts(initDBFilePath),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

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
