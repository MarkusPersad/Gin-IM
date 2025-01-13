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
	GetFriendList(ctx *gin.Context, claims *types.GIClaims) ([]types.Friend, error)
	AddToBlackList(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error
	GetBlackList(ctx *gin.Context, claims *types.GIClaims) ([]types.Friend, error)
	CancelBlack(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error
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

// checkIsFriend 检查两个用户是否已经是朋友关系。
// 参数:
//
//	tx *gorm.DB: 数据库事务对象，用于查询操作。
//	userId, friendId string: 需要检查的两个用户的ID。
//
// 返回值:
//
//	error: 如果用户已经是朋友关系，则返回ErrAlreadyExist错误；否则返回nil。
func checkIsFriend(tx *gorm.DB, userId, friendId string) error {
	var userFriend model.UserFriend
	// 使用userId和friendId进行双向查询，以检查是否存在朋友关系。
	if err := tx.Model(&userFriend).Where("userid = ? And friendid = ?", userId, friendId).Or("userid = ? And friendid = ?", friendId, userId).First(&userFriend).Error; err == nil {
		// 如果查询成功，表示朋友关系已存在，返回ErrAlreadyExist错误。
		return exception.ErrAlreadyExist
	}
	// 如果查询失败，且错误为nil，表示朋友关系不存在，返回nil。
	return nil
}

func (s *service) GetFriendList(ctx *gin.Context, claims *types.GIClaims) ([]types.Friend, error) {
	var friendList []types.Friend
	err := s.Transaction(ctx, func(ctx context.Context) error {
		if err := s.GetDB(ctx).Model(&model.UserFriend{}).Select("user.email, user.username, user.avatar").
			Joins("JOIN user ON user_friend.friendid = user.uuid").
			Where("user_friend.userid = ?", claims.UserId).Scan(&friendList).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return friendList, nil
}

func (s *service) AddToBlackList(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var friend model.User
		if err := s.GetDB(ctx).Model(&model.User{}).Where("username = ?", friendInfo.FriendInfo).Or("email = ?", friendInfo.FriendInfo).First(&friend).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}
		if err := checkIsFriend(s.GetDB(ctx), claims.UserId, friend.Uuid); err == nil {
			return exception.ErrNotFound
		}
		if err := s.GetDB(ctx).Model(&model.UserFriend{}).Where("userid = ? AND friendid = ?", claims.UserId, friend.Uuid).Or("userid = ? AND friendid = ?", friend.Uuid, claims.UserId).Delete(&model.UserFriend{}).Error; err != nil {
			log.Logger.Error().Err(err).Msg("删除失败")
			return err
		}
		return nil
	})
}

func (s *service) GetBlackList(ctx *gin.Context, claims *types.GIClaims) ([]types.Friend, error) {
	var blackList []types.Friend
	err := s.Transaction(ctx, func(ctx context.Context) error {
		if err := s.GetDB(ctx).Unscoped().Model(&model.UserFriend{}).Select("user.email, user.username, user.avatar").
			Joins("JOIN user ON user_friend.friendid = user.uuid").
			Where("user_friend.userid = ? AND user_friend.deleted_at IS NOT NULL", claims.UserId).
			Scan(&blackList).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blackList, nil
}

func (s *service) CancelBlack(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var friend model.User
		if err := s.GetDB(ctx).Model(&model.User{}).Where("username = ?", friendInfo.FriendInfo).Or("email = ?", friendInfo.FriendInfo).First(&friend).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}
		if err := s.GetDB(ctx).Unscoped().Model(&model.UserFriend{}).Where("userid = ? AND friendid = ?", claims.UserId, friend.Uuid).Or("userid = ? AND friendid = ?", friend.Uuid, claims.UserId).Update("deleted_at", nil).Error; err != nil {
			log.Logger.Error().Err(err).Msg("更新失败")
			return err
		}
		return nil
	})
}
