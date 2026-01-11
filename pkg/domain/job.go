package domain

type Job struct {
	CommonModel
	Name string `json:"name" gorm:"name"`
}

func (Job) TableName() string {
	return "m_job"
}
