package types

type Friend struct {
	Email    string `json:"email" gorm:"column:email"`
	Username string `json:"username" gorm:"column:username"`
	Avatar   string `json:"avatar" gorm:"column:avatar"`
	Status   int8   `json:"status" gorm:"column:status"`
}
