package server

import (
	_ "Gin-IM/cmd/api/docs"
	"Gin-IM/internal/midleware"
	"Gin-IM/pkg/response"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"strings"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()

	//Logger() 插件 Recover() 插件
	r.Use(GinLogger(), GinRecovery(true))

	// Gzip Middleware
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	//Timeout Middleware
	r.Use(midleware.TimeoutMiddleware())

	// Error Middleware
	r.Use(midleware.ErrorHandler())

	//JWT Middleware
	r.Use(midleware.JwtMiddleware(func(ctx *gin.Context) bool {
		return strings.Contains(ctx.Request.URL.Path, "/api/account/login") ||
			strings.Contains(ctx.Request.URL.Path, "/api/account/register") ||
			strings.Contains(ctx.Request.URL.Path, "/api/account/getcaptcha") ||
			strings.Contains(ctx.Request.URL.Path, "/hello") ||
			strings.Contains(ctx.Request.URL.Path, "/swagger/")
	}))

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/hello", s.HelloWorldHandler)

	r.GET("/health", s.HealthHandler)
	api := r.Group("/api")
	{
		account := api.Group("/account")
		{
			account.GET("/getcaptcha", s.GetCaptcha)
			account.POST("/register", s.Register)
		}
	}
	return r
}

// HelloWorldHandler godoc
// @Summary Hello World
// @Description Hello World
// @Tags HelloWorld
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /hello [get]
func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"
	c.JSON(http.StatusOK, response.Success(0, "Success", resp))
}
