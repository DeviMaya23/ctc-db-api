package domain

type Traveller struct {
	CommonModel
	Name        string     `json:"name" gorm:"name"`
	Rarity      int        `json:"rarity" gorm:"rarity"`
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
	Name      string                  `json:"name" validate:"required,lte=50"`
	Rarity    int                     `json:"rarity" validate:"required"`
	Influence string                  `json:"influence" validate:"required,influence"`
	Job       string                  `json:"job" validate:"required,job"`
	Accessory *CreateAccessoryRequest `json:"accessory" validate:"omitempty"`
}

type UpdateTravellerRequest struct {
	Name      string                  `json:"name" validate:"required,lte=50"`
	Rarity    int                     `json:"rarity" validate:"required"`
	Influence string                  `json:"influence" validate:"required,influence"`
	Job       string                  `json:"job" validate:"required,job"`
	Accessory *UpdateAccessoryRequest `json:"accessory" validate:"omitempty"`
}

// Response DTOs

type TravellerResponse struct {
	Name      string             `json:"name"`
	Rarity    int                `json:"rarity"`
	Influence string             `json:"influence"`
	Job       string             `json:"job"`
	Accessory *AccessoryResponse `json:"accessory,omitempty"`
}

// Mapper functions

func ToTravellerResponse(traveller Traveller) TravellerResponse {
	return TravellerResponse{
		Name:      traveller.Name,
		Rarity:    traveller.Rarity,
		Influence: traveller.Influence.Name,
		Job:       traveller.Job.Name,
		Accessory: ToAccessoryResponse(traveller.Accessory),
	}
}
