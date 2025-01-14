package server

import (
	"Gin-IM/internal/handler"
	"Gin-IM/internal/model"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int

	*handler.Handlers
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		Handlers: handler.NewHandler(),
	}
	if err := NewServer.InitDBTables(&model.User{}, &model.UserFriend{}, &model.Group{}, &model.File{}); err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to initialize database tables")
	}
	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
