package constants

const (
	InfluenceWealth    = "Wealth"
	InfluencePower     = "Power"
	InfluenceFame      = "Fame"
	InfluenceOpulence  = "Opulence"
	InfluenceDominance = "Dominance"
	InfluencePrestige  = "Prestige"

	InfluenceWealthID    = 1
	InfluencePowerID     = 2
	InfluenceFameID      = 3
	InfluenceOpulenceID  = 4
	InfluenceDominanceID = 5
	InfluencePrestigeID  = 6
)

var (
	influenceMap = map[string]int{
		InfluenceWealth:    InfluenceWealthID,
		InfluencePower:     InfluencePowerID,
		InfluenceFame:      InfluenceFameID,
		InfluenceOpulence:  InfluenceOpulenceID,
		InfluenceDominance: InfluenceDominanceID,
		InfluencePrestige:  InfluencePrestigeID,
	}
)

func GetInfluenceID(influenceName string) int {
	res, exist := influenceMap[influenceName]
	if !exist {
		return 0
	}

	return res
}

const (
	JobWarrior    = "Warrior"
	JobMerchant   = "Merchant"
	JobThief      = "Thief"
	JobApothecary = "Apothecary"
	JobHunter     = "Hunter"
	JobCleric     = "Cleric"
	JobScholar    = "Scholar"
	JobDancer     = "Dancer"

	JobWarriorID    = 1
	JobMerchantID   = 2
	JobThiefID      = 3
	JobApothecaryID = 4
	JobHunterID     = 5
	JobClericID     = 6
	JobScholarID    = 7
	JobDancerID     = 8
)

var (
	jobMap = map[string]int{
		JobWarrior:    JobWarriorID,
		JobMerchant:   JobMerchantID,
		JobThief:      JobThiefID,
		JobApothecary: JobApothecaryID,
		JobHunter:     JobHunterID,
		JobCleric:     JobClericID,
		JobScholar:    JobScholarID,
		JobDancer:     JobDancerID,
	}
)

func GetJobID(jobName string) int {
	res, exist := jobMap[jobName]
	if !exist {
		return 0
	}
	return res
}
