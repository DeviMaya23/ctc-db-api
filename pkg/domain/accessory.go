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
	HP     int    `json:"hp"`
	SP     int    `json:"sp"`
	PAtk   int    `json:"patk"`
	PDef   int    `json:"pdef"`
	EAtk   int    `json:"eatk"`
	EDef   int    `json:"edef"`
	Spd    int    `json:"spd"`
	Crit   int    `json:"crit"`
	Effect string `json:"effect" validate:"omitempty,lte=200"`
}

type UpdateAccessoryRequest struct {
	Name   string `json:"name" validate:"required,lte=50"`
	HP     int    `json:"hp"`
	SP     int    `json:"sp"`
	PAtk   int    `json:"patk"`
	PDef   int    `json:"pdef"`
	EAtk   int    `json:"eatk"`
	EDef   int    `json:"edef"`
	Spd    int    `json:"spd"`
	Crit   int    `json:"crit"`
	Effect string `json:"effect" validate:"omitempty,lte=200"`
}

// Response DTOs

type AccessoryResponse struct {
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	SP     int    `json:"sp"`
	PAtk   int    `json:"patk"`
	PDef   int    `json:"pdef"`
	EAtk   int    `json:"eatk"`
	EDef   int    `json:"edef"`
	Spd    int    `json:"spd"`
	Crit   int    `json:"crit"`
	Effect string `json:"effect"`
}

// Request DTOs

type ListAccessoryRequest struct {
	Owner    string `query:"owner"`
	Effect   string `query:"effect"`
	OrderBy  string `query:"order_by" validate:"omitempty,oneof=hp sp patk pdef eatk edef spd crit"`
	OrderDir string `query:"order_dir" validate:"omitempty,oneof=asc desc"`
}

// AccessoryListItemResponse represents an accessory with its owner's name
// Note: Each accessory can only be owned by one traveller (stored as traveller.accessory_id FK)
type AccessoryListItemResponse struct {
	Name   string `json:"name"`
	HP     int    `json:"hp"`
	SP     int    `json:"sp"`
	PAtk   int    `json:"patk"`
	PDef   int    `json:"pdef"`
	EAtk   int    `json:"eatk"`
	EDef   int    `json:"edef"`
	Spd    int    `json:"spd"`
	Crit   int    `json:"crit"`
	Effect string `json:"effect"`
	Owner  string `json:"owner"`
}

// Mapper functions

func ToAccessoryResponse(accessory *Accessory) *AccessoryResponse {
	if accessory == nil {
		return nil
	}
	return &AccessoryResponse{
		Name:   accessory.Name,
		HP:     accessory.HP,
		SP:     accessory.SP,
		PAtk:   accessory.PAtk,
		PDef:   accessory.PDef,
		EAtk:   accessory.EAtk,
		EDef:   accessory.EDef,
		Spd:    accessory.Spd,
		Crit:   accessory.Crit,
		Effect: accessory.Effect,
	}
}

func ToAccessoryListItemResponse(accessory Accessory, ownerNames map[int64]string) AccessoryListItemResponse {
	return AccessoryListItemResponse{
		Name:   accessory.Name,
		HP:     accessory.HP,
		SP:     accessory.SP,
		PAtk:   accessory.PAtk,
		PDef:   accessory.PDef,
		EAtk:   accessory.EAtk,
		EDef:   accessory.EDef,
		Spd:    accessory.Spd,
		Crit:   accessory.Crit,
		Effect: accessory.Effect,
		Owner:  ownerNames[accessory.ID],
	}
}
