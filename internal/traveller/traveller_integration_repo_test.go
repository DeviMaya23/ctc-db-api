package traveller

import (
	"context"
	"database/sql"
	"errors"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"

	"github.com/stretchr/testify/assert"

	pgGormDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTravellerRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()

	connStr := helpers.GetTestDB(t)

	dbConn, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatal("failed open database ", err)
	}
	db, err := gorm.Open(pgGormDriver.New(pgGormDriver.Config{
		Conn: dbConn,
	}), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		t.Fatal("failed to open gorm ", err)
	}

	logger, _ := logging.NewDevelopmentLogger()

	t.Run("create and retrieve traveller with accessory", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()

		repo := NewTravellerRepository(tx, logger)

		newAcc := &domain.Accessory{Name: "Yusia's Fan", HP: 100, SP: 50}
		assert.Nil(t, tx.WithContext(ctx).Create(newAcc).Error)
		newAccID := int(newAcc.ID)

		assert.Nil(t, repo.Create(ctx, &domain.Traveller{
			Name:        "Celine",
			Rarity:      5,
			InfluenceID: 3,
			JobID:       8,
			AccessoryID: &newAccID,
		}))

		traveller, err := repo.GetByID(ctx, int(newAcc.ID))
		assert.Nil(t, err)
		assert.Equal(t, "Celine", traveller.Name)
		assert.Equal(t, 5, traveller.Rarity)
		assert.Equal(t, "Yusia's Fan", traveller.Accessory.Name)
	})

	t.Run("list travellers with pagination", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()

		repo := NewTravellerRepository(tx, logger)

		assert.Nil(t, tx.WithContext(ctx).Create(&domain.Traveller{Name: "Tahir", Rarity: 4, InfluenceID: 2, JobID: 1}).Error)
		assert.Nil(t, tx.WithContext(ctx).Create(&domain.Traveller{Name: "Celine", Rarity: 5, InfluenceID: 3, JobID: 8}).Error)

		resList, total, err := repo.GetList(ctx, domain.ListTravellerRequest{}, 0, 10)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), total)

		assert.Equal(t, "Tahir", resList[0].Name)
		assert.Equal(t, "Celine", resList[1].Name)
	})

	t.Run("update traveller fields", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()

		repo := NewTravellerRepository(tx, logger)

		tr := &domain.Traveller{Name: "Meena", Rarity: 4, InfluenceID: 1, JobID: 1}
		assert.Nil(t, tx.WithContext(ctx).Create(tr).Error)

		err := repo.Update(ctx, &domain.Traveller{
			CommonModel: domain.CommonModel{ID: tr.ID},
			Rarity:      6,
			Accessory:   &domain.Accessory{Name: "Ribbon"},
		})
		assert.Nil(t, err)

		updated, err := repo.GetByID(ctx, int(tr.ID))
		assert.Nil(t, err)
		assert.Equal(t, 6, updated.Rarity)
		assert.Equal(t, "Ribbon", updated.Accessory.Name)
	})
	t.Run("delete traveller", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()

		repo := NewTravellerRepository(tx, logger)

		tr := &domain.Traveller{Name: "DeleteMe", Rarity: 3, InfluenceID: 1, JobID: 1}
		assert.Nil(t, tx.WithContext(ctx).Create(tr).Error)

		assert.Nil(t, repo.Delete(ctx, int(tr.ID)))

		_, err := repo.GetByID(ctx, int(tr.ID))
		var nfe *domain.NotFoundError
		assert.True(t, errors.As(err, &nfe), "expected NotFoundError but got: %v", err)
	})
}
