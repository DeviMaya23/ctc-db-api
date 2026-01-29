package domain

import (
	"lizobly/ctc-db-api/pkg/constants"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestToTravellerListItemResponse tests mapper function for list item responses
func TestToTravellerListItemResponse(t *testing.T) {
	tests := []struct {
		name      string
		traveller *Traveller
		expected  TravellerListItemResponse
	}{
		{
			name: "complete traveller data",
			traveller: &Traveller{
				Name:        "Alfyn",
				Rarity:      5,
				Banner:      "Summer Banner",
				ReleaseDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				InfluenceID: constants.InfluenceWealthID,
				JobID:       constants.JobApothecaryID,
			},
			expected: TravellerListItemResponse{
				Name:        "Alfyn",
				Rarity:      5,
				Banner:      "Summer Banner",
				ReleaseDate: "15-06-2024",
				Influence:   constants.InfluenceWealth,
				Job:         constants.JobApothecary,
			},
		},
		{
			name: "minimal traveller data",
			traveller: &Traveller{
				Name:        "Ochette",
				Rarity:      3,
				Banner:      "",
				ReleaseDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				InfluenceID: constants.InfluencePowerID,
				JobID:       constants.JobHunterID,
			},
			expected: TravellerListItemResponse{
				Name:        "Ochette",
				Rarity:      3,
				Banner:      "",
				ReleaseDate: "01-01-2023",
				Influence:   constants.InfluencePower,
				Job:         constants.JobHunter,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTravellerListItemResponse(tt.traveller)

			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Rarity, result.Rarity)
			assert.Equal(t, tt.expected.Banner, result.Banner)
			assert.Equal(t, tt.expected.ReleaseDate, result.ReleaseDate)
			assert.Equal(t, tt.expected.Influence, result.Influence)
			assert.Equal(t, tt.expected.Job, result.Job)
		})
	}
}

// TestToTravellerResponse tests mapper function for detailed responses
func TestToTravellerResponse(t *testing.T) {
	tests := []struct {
		name      string
		traveller *Traveller
		validate  func(t *testing.T, result TravellerResponse)
	}{
		{
			name: "traveller with accessory",
			traveller: &Traveller{
				Name:        "Temenos",
				Rarity:      5,
				Banner:      "Cleric Banner",
				ReleaseDate: time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC),
				InfluenceID: constants.InfluenceFameID,
				JobID:       constants.JobClericID,
				Accessory: &Accessory{
					Name:   "Holy Staff",
					HP:     100,
					SP:     50,
					EAtk:   75,
					EDef:   80,
					Effect: "Increases healing",
				},
			},
			validate: func(t *testing.T, result TravellerResponse) {
				assert.Equal(t, "Temenos", result.Name)
				assert.Equal(t, 5, result.Rarity)
				assert.Equal(t, "20-03-2024", result.ReleaseDate)
				assert.Equal(t, constants.InfluenceFame, result.Influence)
				assert.Equal(t, constants.JobCleric, result.Job)
				assert.NotNil(t, result.Accessory)
				assert.Equal(t, "Holy Staff", result.Accessory.Name)
				assert.Equal(t, 100, result.Accessory.HP)
				assert.Equal(t, 50, result.Accessory.SP)
			},
		},
		{
			name: "traveller without accessory",
			traveller: &Traveller{
				Name:        "Hikari",
				Rarity:      4,
				Banner:      "Warrior Banner",
				ReleaseDate: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
				InfluenceID: constants.InfluenceDominanceID,
				JobID:       constants.JobWarriorID,
				Accessory:   nil,
			},
			validate: func(t *testing.T, result TravellerResponse) {
				assert.Equal(t, "Hikari", result.Name)
				assert.Equal(t, 4, result.Rarity)
				assert.Equal(t, "01-12-2023", result.ReleaseDate)
				assert.Equal(t, constants.InfluenceDominance, result.Influence)
				assert.Equal(t, constants.JobWarrior, result.Job)
				assert.Nil(t, result.Accessory)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTravellerResponse(tt.traveller)
			tt.validate(t, result)
		})
	}
}

// TestTraveller_TableName tests table name method
func TestTraveller_TableName(t *testing.T) {
	traveller := Traveller{}
	assert.Equal(t, "m_traveller", traveller.TableName())
}
