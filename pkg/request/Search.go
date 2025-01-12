package request

type UserSearch struct {
	UserInfo string `json:"userInfo" binding:"required" validate:"required" field_error_info:"信息不能为空"`
}
