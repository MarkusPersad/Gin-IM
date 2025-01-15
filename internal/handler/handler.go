package handler

import (
	"Gin-IM/internal/database"
	"Gin-IM/internal/minio"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handlers struct {
	db          database.Service
	minioClient *minio.MinIOStore
}

func NewHandler() *Handlers {
	client := minio.NewClient(false)
	return &Handlers{
		db:          database.New(),
		minioClient: client,
	}
}

func (h *Handlers) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.db.Health())
}

func (h *Handlers) InitDBTables(models ...any) error {
	if len(models) == 0 {
		return nil
	}
	return h.db.GetDB(nil).AutoMigrate(models...)
}
