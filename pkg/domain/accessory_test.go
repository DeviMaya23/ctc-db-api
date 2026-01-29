package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToAccessoryResponse tests mapper function for accessory responses
func TestToAccessoryResponse(t *testing.T) {
	tests := []struct {
		name      string
		accessory *Accessory
		expected  *AccessoryResponse
	}{
		{
			name: "complete accessory data",
			accessory: &Accessory{
				Name:   "Dragon Sword",
				HP:     150,
				SP:     80,
				PAtk:   200,
				PDef:   100,
				EAtk:   50,
				EDef:   75,
				Spd:    120,
				Crit:   15,
				Effect: "Increases fire damage",
			},
			expected: &AccessoryResponse{
				Name:   "Dragon Sword",
				HP:     150,
				SP:     80,
				PAtk:   200,
				PDef:   100,
				EAtk:   50,
				EDef:   75,
				Spd:    120,
				Crit:   15,
				Effect: "Increases fire damage",
			},
		},
		{
			name: "accessory with zero stats",
			accessory: &Accessory{
				Name:   "Basic Ring",
				HP:     0,
				SP:     0,
				PAtk:   0,
				PDef:   0,
				EAtk:   0,
				EDef:   0,
				Spd:    0,
				Crit:   0,
				Effect: "",
			},
			expected: &AccessoryResponse{
				Name:   "Basic Ring",
				HP:     0,
				SP:     0,
				PAtk:   0,
				PDef:   0,
				EAtk:   0,
				EDef:   0,
				Spd:    0,
				Crit:   0,
				Effect: "",
			},
		},
		{
			name:      "nil accessory returns nil",
			accessory: nil,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToAccessoryResponse(tt.accessory)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.HP, result.HP)
			assert.Equal(t, tt.expected.SP, result.SP)
			assert.Equal(t, tt.expected.PAtk, result.PAtk)
			assert.Equal(t, tt.expected.PDef, result.PDef)
			assert.Equal(t, tt.expected.EAtk, result.EAtk)
			assert.Equal(t, tt.expected.EDef, result.EDef)
			assert.Equal(t, tt.expected.Spd, result.Spd)
			assert.Equal(t, tt.expected.Crit, result.Crit)
			assert.Equal(t, tt.expected.Effect, result.Effect)
		})
	}
}

// TestToAccessoryListItemResponse tests mapper function for list responses
func TestToAccessoryListItemResponse(t *testing.T) {
	tests := []struct {
		name       string
		accessory  *Accessory
		ownerNames map[int64]string
		expected   AccessoryListItemResponse
	}{
		{
			name: "accessory with owner",
			accessory: &Accessory{
				CommonModel: CommonModel{ID: 1},
				Name:        "Steel Blade",
				HP:          100,
				SP:          50,
				PAtk:        150,
				PDef:        80,
				EAtk:        30,
				EDef:        40,
				Spd:         100,
				Crit:        10,
				Effect:      "Critical damage boost",
			},
			ownerNames: map[int64]string{
				1: "Ochette",
				2: "Hikari",
			},
			expected: AccessoryListItemResponse{
				Name:   "Steel Blade",
				HP:     100,
				SP:     50,
				PAtk:   150,
				PDef:   80,
				EAtk:   30,
				EDef:   40,
				Spd:    100,
				Crit:   10,
				Effect: "Critical damage boost",
				Owner:  "Ochette",
			},
		},
		{
			name: "accessory without owner",
			accessory: &Accessory{
				CommonModel: CommonModel{ID: 99},
				Name:        "Ancient Amulet",
				HP:          200,
				SP:          100,
				PAtk:        50,
				PDef:        50,
				EAtk:        150,
				EDef:        150,
				Spd:         80,
				Crit:        5,
				Effect:      "Elemental mastery",
			},
			ownerNames: map[int64]string{
				1: "Ochette",
				2: "Hikari",
			},
			expected: AccessoryListItemResponse{
				Name:   "Ancient Amulet",
				HP:     200,
				SP:     100,
				PAtk:   50,
				PDef:   50,
				EAtk:   150,
				EDef:   150,
				Spd:    80,
				Crit:   5,
				Effect: "Elemental mastery",
				Owner:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToAccessoryListItemResponse(tt.accessory, tt.ownerNames)

			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.HP, result.HP)
			assert.Equal(t, tt.expected.SP, result.SP)
			assert.Equal(t, tt.expected.PAtk, result.PAtk)
			assert.Equal(t, tt.expected.PDef, result.PDef)
			assert.Equal(t, tt.expected.EAtk, result.EAtk)
			assert.Equal(t, tt.expected.EDef, result.EDef)
			assert.Equal(t, tt.expected.Spd, result.Spd)
			assert.Equal(t, tt.expected.Crit, result.Crit)
			assert.Equal(t, tt.expected.Effect, result.Effect)
			assert.Equal(t, tt.expected.Owner, result.Owner)
		})
	}
}

// TestAccessory_TableName tests table name method
func TestAccessory_TableName(t *testing.T) {
	accessory := Accessory{}
	assert.Equal(t, "m_accessory", accessory.TableName())
}
