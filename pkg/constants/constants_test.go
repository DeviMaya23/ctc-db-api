package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInfluenceID(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		want     int
	}{
		{"Wealth", InfluenceWealth, InfluenceWealthID},
		{"Power", InfluencePower, InfluencePowerID},
		{"Fame", InfluenceFame, InfluenceFameID},
		{"Opulence", InfluenceOpulence, InfluenceOpulenceID},
		{"Dominance", InfluenceDominance, InfluenceDominanceID},
		{"Prestige", InfluencePrestige, InfluencePrestigeID},
		{"invalid influence", "InvalidInfluence", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetInfluenceID(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetInfluenceName(t *testing.T) {
	tests := []struct {
		testName string
		input    int
		want     string
	}{
		{"Wealth ID", InfluenceWealthID, InfluenceWealth},
		{"Power ID", InfluencePowerID, InfluencePower},
		{"Fame ID", InfluenceFameID, InfluenceFame},
		{"Opulence ID", InfluenceOpulenceID, InfluenceOpulence},
		{"Dominance ID", InfluenceDominanceID, InfluenceDominance},
		{"Prestige ID", InfluencePrestigeID, InfluencePrestige},
		{"invalid ID", 999, ""},
		{"zero ID", 0, ""},
		{"negative ID", -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetInfluenceName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetJobID(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		want     int
	}{
		{"Warrior", JobWarrior, JobWarriorID},
		{"Merchant", JobMerchant, JobMerchantID},
		{"Thief", JobThief, JobThiefID},
		{"Apothecary", JobApothecary, JobApothecaryID},
		{"Hunter", JobHunter, JobHunterID},
		{"Cleric", JobCleric, JobClericID},
		{"Scholar", JobScholar, JobScholarID},
		{"Dancer", JobDancer, JobDancerID},
		{"invalid job", "InvalidJob", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetJobID(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetJobName(t *testing.T) {
	tests := []struct {
		testName string
		input    int
		want     string
	}{
		{"Warrior ID", JobWarriorID, JobWarrior},
		{"Merchant ID", JobMerchantID, JobMerchant},
		{"Thief ID", JobThiefID, JobThief},
		{"Apothecary ID", JobApothecaryID, JobApothecary},
		{"Hunter ID", JobHunterID, JobHunter},
		{"Cleric ID", JobClericID, JobCleric},
		{"Scholar ID", JobScholarID, JobScholar},
		{"Dancer ID", JobDancerID, JobDancer},
		{"invalid ID", 999, ""},
		{"zero ID", 0, ""},
		{"negative ID", -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetJobName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConstantsValues(t *testing.T) {
	t.Run("date format", func(t *testing.T) {
		assert.Equal(t, "02-01-2006", DateFormat)
	})

	t.Run("order direction", func(t *testing.T) {
		assert.Equal(t, "asc", OrderDirAsc)
		assert.Equal(t, "desc", OrderDirDesc)
	})

	t.Run("cache max age", func(t *testing.T) {
		assert.Equal(t, 300, CacheMaxAgeList)
		assert.Equal(t, 600, CacheMaxAgeResource)
	})
}

func TestInfluenceJobRoundTrip(t *testing.T) {
	t.Run("influence round trip", func(t *testing.T) {
		influences := []string{
			InfluenceWealth, InfluencePower, InfluenceFame,
			InfluenceOpulence, InfluenceDominance, InfluencePrestige,
		}

		for _, influence := range influences {
			id := GetInfluenceID(influence)
			assert.NotZero(t, id, "influence %s should have non-zero ID", influence)

			name := GetInfluenceName(id)
			assert.Equal(t, influence, name, "round trip should return original influence")
		}
	})

	t.Run("job round trip", func(t *testing.T) {
		jobs := []string{
			JobWarrior, JobMerchant, JobThief, JobApothecary,
			JobHunter, JobCleric, JobScholar, JobDancer,
		}

		for _, job := range jobs {
			id := GetJobID(job)
			assert.NotZero(t, id, "job %s should have non-zero ID", job)

			name := GetJobName(id)
			assert.Equal(t, job, name, "round trip should return original job")
		}
	})
}
