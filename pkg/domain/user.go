package domain

type User struct {
	CommonModel
	Username string `json:"username" gorm:"username"`
	Password string `json:"password" gorm:"password"`
	Token    string `json:"token" gorm:"token"`
}

func (User) TableName() string {
	return "m_user"
}

type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"password123"`
}

type LoginResponse struct {
	Username string `json:"username" example:"admin"`
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
