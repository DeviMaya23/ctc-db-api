package postgres

import (
	"context"
	"database/sql"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"

	"github.com/stretchr/testify/assert"

	pgGormDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTravellerRepository_Integration(t *testing.T) {
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
