package response

import (
	"Gin-IM/pkg/exception"
	"errors"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(code int, msg string, data interface{}) Response {
	return Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func Fail(err error) Response {
	var value *exception.PersonalError
	if errors.As(err, &value) {
		return Response{
			Code: value.Code,
			Msg:  value.Error(),
			Data: nil,
		}
	}
	return Response{
		Code: 500,
		Msg:  "内部服务错误",
		Data: nil,
	}
}
