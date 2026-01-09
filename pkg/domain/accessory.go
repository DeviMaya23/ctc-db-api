package domain

type Accessory struct {
	CommonModel
	Name   string `json:"name" gorm:"column:name"`
	HP     int    `json:"hp" gorm:"column:hp"`
	SP     int    `json:"sp" gorm:"column:sp"`
	PAtk   int    `json:"patk" gorm:"column:patk"`
	PDef   int    `json:"pdef" gorm:"column:pdef"`
	EAtk   int    `json:"eatk" gorm:"column:eatk"`
	EDef   int    `json:"edef" gorm:"column:edef"`
	Spd    int    `json:"spd" gorm:"column:spd"`
	Crit   int    `json:"crit" gorm:"column:crit"`
	Effect string `json:"effect" gorm:"column:effect"`
}

func (Accessory) TableName() string {
	return "m_accessory"
}

type CreateAccessoryRequest struct {
	Name   string `json:"name" validate:"required,lte=50"`
	HP     int    `json:"hp" validate:"omitempty,gte=0"`
	SP     int    `json:"sp" validate:"omitempty,gte=0"`
	PAtk   int    `json:"patk" validate:"omitempty,gte=0"`
	PDef   int    `json:"pdef" validate:"omitempty,gte=0"`
	EAtk   int    `json:"eatk" validate:"omitempty,gte=0"`
	EDef   int    `json:"edef" validate:"omitempty,gte=0"`
	Spd    int    `json:"spd" validate:"omitempty,gte=0"`
	Crit   int    `json:"crit" validate:"omitempty,gte=0"`
	Effect string `json:"effect" validate:"omitempty,lte=200"`
}
