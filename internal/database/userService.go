package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/enums"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/token"
	"Gin-IM/pkg/types"
	"Gin-IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"time"
)

type UserService interface {
	Register(ctx context.Context, register request.Register) error

	Login(ctx *gin.Context, login request.Login) (string, error)
}

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

// Login 用户登录函数
// 该函数接收一个登录请求，包含用户邮箱和密码，验证用户信息并返回JWT令牌
// 参数:
//
//	ctx *gin.Context: Gin框架的上下文，用于处理HTTP请求和响应
//	login request.Login: 登录请求对象，包含用户邮箱和密码
//
// 返回值:
//
//	string: 成功登录后返回的JWT令牌
//	error: 登录过程中可能遇到的错误，如果用户不存在、密码错误或数据库操作失败等
func (s *service) Login(ctx *gin.Context, login request.Login) (string, error) {
	var user model.User
	// 使用事务处理登录过程中的数据库操作
	err := s.Transaction(ctx, func(ctx context.Context) error {
		// 查询用户邮箱是否存在于数据库中
		if err := s.GetDB(ctx).Model(&user).Where("email = ?", login.Email).First(&user).Error; err != nil {
			return exception.ErrNotFound
		}
		// 验证用户密码是否正确
		if !utils.CompareHashPassword(user.Password, login.Password) {
			return exception.ErrPassword
		}
		// 更新用户状态为登录状态
		if err := s.GetDB(ctx).Model(&user).Where("uuid = ?", user.Uuid).Update("status", enums.LogIn).Error; err != nil {
			log.Logger.Error().Err(err).Msg("更新用户状态失败")
			return err
		}
		// 设置用户token到缓存中
		if err := s.SetAndTime(ctx, defines.USER_TOKEN_KEY+user.Uuid, user.Uuid+user.Email, defines.USER_TOKEN); err != nil {
			log.Logger.Error().Err(err).Msg("设置用户token失败")
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	// 生成JWT令牌
	claims := types.GIClaims{
		UserId: user.Uuid,
		Admin:  utils.IsAdmin(user.Email),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(time.Hour * defines.TOKEN_EXPIRE),
			},
		},
	}
	tokenString := token.GernerateToken(claims)
	return tokenString, nil
}
