package handler

import "Gin-IM/internal/database"

type Handlers struct {
	db database.Service
}

func NewHandler() *Handlers {
	return &Handlers{
		db: database.New(),
	}
}
