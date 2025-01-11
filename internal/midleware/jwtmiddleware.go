package midleware

import (
	"Gin-IM/pkg/token"
	"github.com/gin-gonic/gin"
)

// JwtMiddleware 返回一个基于JWT的认证中间件，它可以根据条件跳过认证过程。
// skipper 是一个决定是否跳过JWT认证的函数。如果skipper为nil，或者skipper函数调用时返回false，则会进行JWT认证。
func JwtMiddleware(skipper func(c *gin.Context) bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 如果skipper不为空且skipper函数调用时返回true，则跳过JWT认证。
		if skipper != nil && skipper(ctx) {
			ctx.Next()
			return
		}

		// 进行JWT认证，如果认证失败，则终止请求处理并返回错误信息。
		if err := token.TokenValid(ctx); err != nil {
			err = ctx.Error(err)
			ctx.Abort()
			return
		}

		// 认证成功，继续执行下一个中间件或处理函数。
		ctx.Next()
	}
}
