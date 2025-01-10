package midleware

import (
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/exception"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func timeoutResponse(ctx *gin.Context) {
	ctx.JSON(http.StatusRequestTimeout, exception.ErrTimeout)
}

func TimeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(defines.Timeout*time.Millisecond),
		timeout.WithHandler(func(ctx *gin.Context) {
			ctx.Next()
		}),
		timeout.WithResponse(timeoutResponse),
	)
}
