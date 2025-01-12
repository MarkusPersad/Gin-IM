package request

type FriendRequest struct {
	FriendInfo string `json:"friendInfo" binding:"required" validate:"required" field_error_info:"好友信息不能为空"`
}
