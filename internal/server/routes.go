package server

import (
	_ "Gin-IM/cmd/api/docs"
	"Gin-IM/internal/midleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()

	//Logger() 插件 Recover() 插件
	r.Use(GinLogger(), GinRecovery(true))

	//Timeout Middleware

	r.Use(midleware.TimeoutMiddleware())

	// Error Middleware
	r.Use(midleware.ErrorHandler())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	return r
}

// HelloWorldHandler godoc
// @Summary Hello World
// @Description Hello World
// @Tags HelloWorld
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Router / [get]
func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"
	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
