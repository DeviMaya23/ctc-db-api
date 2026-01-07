package postgres

import (
	"context"
	"database/sql"
	"lizobly/cotc-db-api/pkg/domain"
	"lizobly/cotc-db-api/pkg/logging"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	pgGormDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTravellerRepository_Integration(t *testing.T) {
	ctx := context.Background()

	initDBFilePath := filepath.Join("../../..", "testdata", "db-traveller-repo.sql")

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
	repo := NewTravellerRepository(db, logger)

	errCreate := repo.Create(ctx, &domain.Traveller{
		Name:        "Fiore",
		Rarity:      5,
		InfluenceID: 3,
	})
	assert.Nil(t, errCreate)

	traveller, err := repo.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, traveller.Name, "Fiore")
	assert.Equal(t, traveller.Rarity, 5)
	assert.Equal(t, traveller.InfluenceID, 3)

	// Update traveller
	err = repo.Update(ctx, &domain.Traveller{
		CommonModel: domain.CommonModel{
			ID: 1,
		},
		Rarity: 6,
	})
	assert.Nil(t, err)

	// Check updated traveller
	traveller, err = repo.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, traveller.Name, "Fiore")
	assert.Equal(t, traveller.Rarity, 6)

	// Delete traveller
	err = repo.Delete(ctx, 1)
	assert.Nil(t, err)

	// Check if deleted
	traveller, err = repo.GetByID(ctx, 1)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

}
