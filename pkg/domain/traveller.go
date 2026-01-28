package domain

import (
	"lizobly/ctc-db-api/pkg/constants"
	"time"
)

type Traveller struct {
	CommonModel
	Name        string     `json:"name" gorm:"name"`
	Rarity      int        `json:"rarity" gorm:"rarity"`
	Banner      string     `json:"banner" gorm:"banner"`
	ReleaseDate time.Time  `json:"release_date" gorm:"release_date"`
	InfluenceID int        `json:"influence_id" gorm:"influence_id"`
	Influence   Influence  `json:"influence" gorm:"foreignKey:influence_id"`
	JobID       int        `json:"job_id" gorm:"job_id"`
	Job         Job        `json:"job" gorm:"foreignKey:job_id"`
	AccessoryID *int       `json:"-" gorm:"accessory_id"`
	Accessory   *Accessory `json:"accessory,omitempty" gorm:"foreignKey:accessory_id"`
}

func (Traveller) TableName() string {
	return "m_traveller"
}

type CreateTravellerRequest struct {
	Name        string                  `json:"name" validate:"required,lte=50"`
	Rarity      int                     `json:"rarity" validate:"required,gte=1,lte=5"`
	Banner      string                  `json:"banner" validate:"omitempty,lte=50"`
	ReleaseDate string                  `json:"release_date" validate:"omitempty,datetime=02-01-2006"`
	Influence   string                  `json:"influence" validate:"required,influence"`
	Job         string                  `json:"job" validate:"required,job"`
	Accessory   *CreateAccessoryRequest `json:"accessory" validate:"omitempty"`
}

type UpdateTravellerRequest struct {
	Name        string                  `json:"name" validate:"required,lte=50"`
	Rarity      int                     `json:"rarity" validate:"required,gte=1,lte=5"`
	Banner      string                  `json:"banner" validate:"omitempty,lte=50"`
	ReleaseDate string                  `json:"release_date" validate:"omitempty,datetime=02-01-2006"`
	Influence   string                  `json:"influence" validate:"required,influence"`
	Job         string                  `json:"job" validate:"required,job"`
	Accessory   *UpdateAccessoryRequest `json:"accessory" validate:"omitempty"`
}

// Request DTOs

type ListTravellerRequest struct {
	Name        string `query:"name"`
	Influence   string `query:"influence" validate:"omitempty,influence" json:"-"`
	Job         string `query:"job" validate:"omitempty,job" json:"-"`
	InfluenceID int    `json:"-"`
	JobID       int    `json:"-"`
}

// Response DTOs

type TravellerListItemResponse struct {
	Name        string `json:"name"`
	Rarity      int    `json:"rarity"`
	Banner      string `json:"banner"`
	ReleaseDate string `json:"release_date"`
	Influence   string `json:"influence"`
	Job         string `json:"job"`
}

type TravellerResponse struct {
	Name        string             `json:"name"`
	Rarity      int                `json:"rarity"`
	Banner      string             `json:"banner"`
	ReleaseDate string             `json:"release_date"`
	Influence   string             `json:"influence"`
	Job         string             `json:"job"`
	Accessory   *AccessoryResponse `json:"accessory,omitempty"`
}

// Mapper functions

func ToTravellerListItemResponse(traveller *Traveller) TravellerListItemResponse {
	return TravellerListItemResponse{
		Name:        traveller.Name,
		Rarity:      traveller.Rarity,
		Banner:      traveller.Banner,
		ReleaseDate: traveller.ReleaseDate.Format("02-01-2006"),
		Influence:   constants.GetInfluenceName(traveller.InfluenceID),
		Job:         constants.GetJobName(traveller.JobID),
	}
}

func ToTravellerResponse(traveller *Traveller) TravellerResponse {
	return TravellerResponse{
		Name:        traveller.Name,
		Rarity:      traveller.Rarity,
		Banner:      traveller.Banner,
		ReleaseDate: traveller.ReleaseDate.Format("02-01-2006"),
		Influence:   constants.GetInfluenceName(traveller.InfluenceID),
		Job:         constants.GetJobName(traveller.JobID),
		Accessory:   ToAccessoryResponse(traveller.Accessory),
	}
}
