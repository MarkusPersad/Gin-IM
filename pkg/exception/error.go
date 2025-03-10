package exception

var (
	ErrTimeout      = NewError(1000, "请求超时")
	ErrCheckCode    = NewError(1001, "验证码错误")
	ErrInvalidToken = NewError(1002, "Token无效")
	ErrTokenEmpty   = NewError(1003, "Token为空")
	ErrUnknownAlg   = NewError(1004, "未知的加密算法")
	ErrBadRequest   = NewError(1006, "请求参数错误")
	ErrAlreadyExist = NewError(1007, "数据已存在")
	ErrNotFound     = NewError(1008, "数据不存在")
	ErrPassword     = NewError(1009, "密码错误")
	ErrAlreadyLogin = NewError(1010, "用户已登录")
	ErrLoginTimeout = NewError(1011, "登录超时")
	ErrUploadFile   = NewError(1012, "上传文件失败")
	ErrFileUrl      = NewError(1013, "文件链接获取失败")
	//ErrPermissionDenied = NewError(1014, "权限不足")
	ErrFileDelete    = NewError(1015, "文件删除失败")
	ErrFileUploading = NewError(1016, "文件还还不能合并")
	ErrFileRecovery  = NewError(1017, "文件未能恢复")
)

type PersonalError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *PersonalError) Error() string {
	return e.Msg
}

func NewError(code int, msg string) *PersonalError {
	return &PersonalError{
		Code: code,
		Msg:  msg,
	}
}
