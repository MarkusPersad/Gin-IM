package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/utils"
	"context"
	"github.com/google/uuid"
)

func (s *service) Register(ctx context.Context, register request.Register) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var user model.User
		if err := s.GetDB(ctx).Model(&user).Where("email = ? or username = ?", register.Email, register.UserName).First(&user).Error; err == nil {
			return exception.ErrAlreadyExist
		}
		user.Uuid = uuid.New().String()
		user.Email = register.Email
		user.Username = register.UserName
		user.Password = utils.GernerateHashPassword(register.Password)
		if err := s.GetDB(ctx).Create(&user).Error; err != nil {
			return err
		}
		return nil
	})
}
