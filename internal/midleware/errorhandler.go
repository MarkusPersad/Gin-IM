package midleware

import (
	"Gin-IM/pkg/exception"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) == 0 {
			return
		}
		if err, ok := ctx.Errors.Last().Err.(*exception.PersonalError); ok {
			ctx.JSON(http.StatusOK, err)
			ctx.Abort()
			return
		}
		log.Logger.Error().Err(ctx.Errors.Last()).Msg("Unkown ErrorHandler")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "内部服务错误",
		})
	}
}
