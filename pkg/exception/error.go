package exception

var (
	ErrTimeout = NewError(1000, "请求超时")
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
