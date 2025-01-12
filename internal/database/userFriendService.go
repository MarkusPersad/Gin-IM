package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/types"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type UserFriendService interface {
	AddFriend(ctx *gin.Context, claims *types.GIClaims, request request.FriendRequest) error
}

func (s *service) AddFriend(ctx *gin.Context, claims *types.GIClaims, request request.FriendRequest) error {
	err := s.Transaction(ctx, func(ctx context.Context) error {
		var friend model.User
		if err := s.GetDB(ctx).Model(&friend).Where("username = ?", request.FriendInfo).Or("email = ?", request.FriendInfo).First(&friend).Error; err != nil {
			return exception.ErrNotFound
		}
		if err := checkIsFriend(s.GetDB(ctx), claims.UserId, friend.Uuid); err != nil {
			return err
		}
		var userFriend model.UserFriend
		userFriend.UserId = claims.UserId
		userFriend.FriendId = friend.Uuid
		if err := s.GetDB(ctx).Create(&userFriend).Error; err != nil {
			log.Logger.Error().Err(err).Msg("插入失败")
			return err
		}
		var friendUser model.UserFriend
		friendUser.UserId = friend.Uuid
		friendUser.FriendId = claims.UserId
		if err := s.GetDB(ctx).Create(&friendUser).Error; err != nil {
			log.Logger.Error().Err(err).Msg("插入失败")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func checkIsFriend(tx *gorm.DB, userId, friendId string) error {
	var userFriend model.UserFriend
	if err := tx.Model(&userFriend).Where("userid = ? And friendid = ?", userId, friendId).Or("userid = ? And friendid = ?", friendId, userId).First(&userFriend).Error; err == nil {
		return exception.ErrAlreadyExist
	}
	return nil
}
