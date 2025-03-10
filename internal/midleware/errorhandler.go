package midleware

import (
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/response"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) == 0 {
			return
		}
		var err *exception.PersonalError
		if errors.As(ctx.Errors.Last().Err, &err) {
			ctx.JSON(http.StatusOK, response.Fail(err))
			ctx.Abort()
			return
		}
		ctx.JSON(http.StatusInternalServerError, response.Fail(ctx.Errors.Last().Err))
	}
}
