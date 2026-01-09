package domain

type Traveller struct {
	CommonModel
	Name        string     `json:"name" gorm:"name"`
	Rarity      int        `json:"rarity" gorm:"rarity"`
	InfluenceID int        `json:"influence_id" gorm:"influence_id"`
	Influence   Influence  `json:"influence" gorm:"foreignKey:influence_id"`
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
	Accessory *CreateAccessoryRequest `json:"accessory" validate:"omitempty"`
}

type UpdateTravellerRequest struct {
	Name      string                  `json:"name" validate:"required,lte=50"`
	Rarity    int                     `json:"rarity" validate:"required"`
	Influence string                  `json:"influence" validate:"required,influence"`
	Accessory *UpdateAccessoryRequest `json:"accessory" validate:"omitempty"`
}
