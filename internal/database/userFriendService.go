package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/enums"
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
	DeleteFriend(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error
	AgreeFriendRequest(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error
}

// AddFriend 添加好友
// 该函数通过事务处理来添加双向好友关系
// 参数:
//
//	ctx *gin.Context: Gin框架的上下文对象，用于处理HTTP请求和响应
//	claims *types.GIClaims: 包含用户信息的令牌声明，用于获取当前用户ID
//	request request.FriendRequest: 包含好友请求信息的结构体，此处主要用于获取好友的用户名或邮箱
//
// 返回值:
//
//	error: 如果执行过程中出现错误，则返回相应的错误
func (s *service) AddFriend(ctx *gin.Context, claims *types.GIClaims, request request.FriendRequest) error {
	// 使用事务处理来确保数据一致性
	err := s.Transaction(ctx, func(ctx context.Context) error {
		// 查询好友信息，根据提供的用户名或邮箱定位用户
		var friend model.User
		if err := s.GetDB(ctx).Model(&friend).Where("username = ?", request.FriendInfo).Or("email = ?", request.FriendInfo).First(&friend).Error; err != nil {
			// 如果找不到指定的好友，则返回未找到错误
			return exception.ErrNotFound
		}
		// 检查当前用户与目标用户是否已经是好友
		if err := checkIsFriend(s.GetDB(ctx), claims.UserId, friend.Uuid); err != nil {
			// 如果已经是好友或者出现其他错误，则返回相应的错误
			return err
		}
		// 创建用户与好友的关系记录
		var userFriend model.UserFriend
		userFriend.UserId = claims.UserId
		userFriend.FriendId = friend.Uuid
		if err := s.GetDB(ctx).Create(&userFriend).Error; err != nil {
			// 如果插入失败，则记录错误日志并返回错误
			log.Logger.Error().Err(err).Msg("插入失败")
			return err
		}
		// 创建好友与用户的关系记录，确保双向关系
		var friendUser model.UserFriend
		friendUser.UserId = friend.Uuid
		friendUser.FriendId = claims.UserId
		if err := s.GetDB(ctx).Create(&friendUser).Error; err != nil {
			// 如果插入失败，则记录错误日志并返回错误
			log.Logger.Error().Err(err).Msg("插入失败")
			return err
		}
		// 事务执行成功，返回nil
		return nil
	})
	// 如果事务处理过程中出现错误，则返回错误
	if err != nil {
		return err
	}
	// 函数执行成功，返回nil
	return nil
}

// checkIsFriend 检查两个用户是否已经是朋友关系
// 参数:
//
//	tx *gorm.DB: 数据库事务对象
//	userId string: 用户A的ID
//	friendId string: 用户B的ID
//
// 返回值:
//
//	error: 如果用户已经是朋友关系，则返回异常错误；否则返回nil
func checkIsFriend(tx *gorm.DB, userId, friendId string) error {
	var userFriend model.UserFriend
	// 查询用户之间的朋友关系记录，考虑两种情况：(用户A是发起者，用户B是接收者)或(用户B是发起者，用户A是接收者)
	// 并且排除不是朋友关系的状态
	if err := tx.Model(&model.UserFriend{}).
		Where("userid = ? And friendid = ?", userId, friendId).
		Or("userid = ? And friendid = ?", friendId, userId).
		First(&userFriend).Error; err == nil {
		// 如果查询成功，表示用户已经是朋友关系，返回异常错误
		return exception.ErrAlreadyExist
	}
	// 如果查询失败，表示用户还不是朋友关系，返回nil
	return nil
}

// GetFriendList 获取用户的好友列表
// 该方法使用了事务来确保数据的一致性，通过用户的ID来查询好友信息，并排除了不是好友的状态
// 参数:
//
//	ctx *gin.Context - Gin框架的上下文，用于处理HTTP请求和响应
//	claims *types.GIClaims - 包含用户信息的令牌声明，用于获取用户ID
//
// 返回值:
//
//	[]types.Friend - 好友列表，包含好友的邮箱、用户名和头像
//	error - 错误信息，如果执行成功则为nil
func (s *service) GetFriendList(ctx *gin.Context, claims *types.GIClaims) ([]types.Friend, error) {
	var friendList []types.Friend
	// 使用事务处理，确保数据查询的一致性和完整性
	err := s.Transaction(ctx, func(ctx context.Context) error {
		// 通过用户ID查询好友信息，排除不是好友的状态
		if err := s.GetDB(ctx).Model(&model.UserFriend{}).Select("user.email, user.username, user.avatar,user_friend.status").
			Joins("JOIN user ON user_friend.friendid = user.uuid").
			Where("user_friend.userid = ? AND user_friend.status != ?", claims.UserId, enums.NOT_FRIEND).Scan(&friendList).Error; err != nil {
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

// AddToBlackList 将指定的用户添加到当前用户的黑名单中。
//
// 该函数执行以下操作：
// 1. 根据用户名或邮箱查询目标用户信息。
// 2. 如果找到目标用户，删除与该用户的好友关系（软删除）。
// 3. 更新已删除的好友关系状态为黑名单状态。
//
// 参数:
//
//	ctx *gin.Context: Gin框架的上下文，用于处理HTTP请求和响应。
//	claims *types.GIClaims: 包含当前用户信息的令牌声明。
//	friendInfo request.FriendRequest: 包含要添加到黑名单的用户信息的请求对象。
//
// 返回值:
//
//	error: 如果操作失败，返回相应的错误。如果成功，则返回nil。
func (s *service) AddToBlackList(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var friend model.User

		// 查询指定的用户，根据用户名或邮箱进行查找
		if err := s.GetDB(ctx).Model(&model.User{}).Where("username = ?", friendInfo.FriendInfo).Or("email = ?", friendInfo.FriendInfo).First(&friend).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}

		// 删除与指定用户的好友关系（软删除），如果存在好友关系
		if err := s.GetDB(ctx).Model(&model.UserFriend{}).
			Where("userid = ? AND friendid = ?", claims.UserId, friend.Uuid).
			Or("userid = ? AND friendid = ?", friend.Uuid, claims.UserId).
			Where("status = ?", enums.IS_FRIEND).
			Delete(&model.UserFriend{}).Error; err != nil {
			log.Logger.Error().Err(err).Msg("删除失败")
			return err
		}
		if err := s.GetDB(ctx).Unscoped().Model(&model.UserFriend{}).
			Where("userid = ? AND friendid = ?", claims.UserId, friend.Uuid).
			Where("status = ?", enums.IS_FRIEND).
			Update("status", enums.BLACK).Error; err != nil {
			log.Logger.Error().Err(err).Msg("更新失败")
			return err
		}
		// 更新已软删除的好友关系状态为黑名单状态
		if err := s.GetDB(ctx).Unscoped().Model(&model.UserFriend{}).
			Where("userid = ? AND friendid = ?", friend.Uuid, claims.UserId).
			Where("status = ?", enums.IS_FRIEND).
			Update("status", enums.BLACKED).Error; err != nil {
			log.Logger.Error().Err(err).Msg("更新失败")
			return err
		}
		return nil
	})
}

// GetBlackList 获取用户的黑名单列表
// 该方法通过用户信息（claims）查询并返回用户的黑名单列表
// 参数:
//
//	ctx *gin.Context - Gin框架的上下文，用于处理HTTP请求和响应
//	claims *types.GIClaims - 用户的令牌信息，包含用户ID等数据
//
// 返回值:
//
//	[]types.Friend - 黑名单列表，包含用户邮箱、用户名和头像等信息
//	error - 如果查询失败，返回错误信息
func (s *service) GetBlackList(ctx *gin.Context, claims *types.GIClaims) ([]types.Friend, error) {
	var blackList []types.Friend
	// 使用事务处理黑名单查询，确保数据一致性
	err := s.Transaction(ctx, func(ctx context.Context) error {
		// 查询已删除的用户好友关系，获取黑名单列表
		if err := s.GetDB(ctx).Unscoped().Model(&model.UserFriend{}).
			Select("user.email, user.username, user.avatar,user_friend.status").
			Joins("JOIN user ON user_friend.friendid = user.uuid").
			Where("user_friend.userid = ?", claims.UserId).
			Where("user_friend.deleted_at IS NOT NULL").
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

// CancelBlack 取消黑名单操作。
// 该函数通过用户名或邮箱查找用户，并取消与其的黑名单关系，转而建立正常好友关系。
// 参数:
//
//	ctx *gin.Context: HTTP请求上下文。
//	claims *types.GIClaims: 包含用户信息的令牌声明。
//	friendInfo request.FriendRequest: 包含朋友信息的请求体。
//
// 返回值:
//
//	error: 如果操作失败，返回相应的错误。
func (s *service) CancelBlack(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var friend model.User
		// 查询朋友信息，根据用户名或邮箱获取用户记录。
		if err := s.GetDB(ctx).Model(&model.User{}).Where("username = ?", friendInfo.FriendInfo).Or("email = ?", friendInfo.FriendInfo).First(&friend).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}
		// 更新UserFriend表，恢复与朋友的关联，并设置为正常好友状态。
		if err := s.GetDB(ctx).Unscoped().Model(&model.UserFriend{}).
			Where("userid = ? AND friendid = ?", claims.UserId, friend.Uuid).
			Or("userid = ? AND friendid = ?", friend.Uuid, claims.UserId).
			Where("deleted_at IS NOT NULL").
			Updates(map[string]interface{}{"deleted_at": nil, "status": enums.IS_FRIEND}).Error; err != nil {
			log.Logger.Error().Err(err).Msg("更新失败")
			return err
		}
		return nil
	})
}

// DeleteFriend 删除用户的好友
// 该函数通过用户名或邮箱查找好友的用户信息，然后删除双方的好友关系
// 参数:
//
//	ctx *gin.Context: Gin框架的上下文对象，用于处理HTTP请求和响应
//	claims *types.GIClaims: 包含用户信息的令牌声明，用于获取当前用户ID
//	friendInfo request.FriendRequest: 包含好友信息的请求对象，用于查找好友
//
// 返回值:
//
//	error: 如果删除操作失败，返回相应的错误
func (s *service) DeleteFriend(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var friend model.User
		// 根据用户名或邮箱查询好友信息，如果查询失败，返回NotFound错误
		if err := s.GetDB(ctx).Model(&model.User{}).Where("username = ?", friendInfo.FriendInfo).Or("email = ?", friendInfo.FriendInfo).First(&friend).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}
		// 删除用户和好友之间的关系，如果删除失败，返回错误
		if err := s.GetDB(ctx).Unscoped().Model(&model.UserFriend{}).Where("userid = ? AND friendid = ?", claims.UserId, friend.Uuid).Or("userid = ? AND friendid = ?", friend.Uuid, claims.UserId).Delete(&model.UserFriend{}).Error; err != nil {
			log.Logger.Error().Err(err).Msg("删除失败")
			return err
		}
		return nil
	})
}

// AgreeFriendRequest 同意好友请求
// 该函数处理用户同意好友请求的逻辑，主要执行以下操作：
// 1. 根据用户名或邮箱查询好友信息
// 2. 更新用户与好友之间的关系状态为好友
// 参数:
//
//	ctx *gin.Context: HTTP请求上下文
//	claims *types.GIClaims: 用户令牌信息
//	friendInfo request.FriendRequest: 好友请求信息
//
// 返回值:
//
//	error: 错误信息，如果执行成功则为nil
func (s *service) AgreeFriendRequest(ctx *gin.Context, claims *types.GIClaims, friendInfo request.FriendRequest) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var friend model.User
		// 查询好友信息，根据提供的用户名或邮箱定位用户
		if err := s.GetDB(ctx).Model(&model.User{}).
			Where("username = ?", friendInfo.FriendInfo).
			Or("email = ?", friendInfo.FriendInfo).
			First(&friend).Error; err != nil {
			log.Logger.Error().Err(err).Msg("查询失败")
			return exception.ErrNotFound
		}
		// 更新用户关系状态为好友，无论谁先添加对方，都更新双方的关系状态
		if err := s.GetDB(ctx).Model(&model.UserFriend{}).
			Where("userid = ? AND friendid =?", claims.UserId, friend.Uuid).
			Or("userid = ? AND friendid =?", friend.Uuid, claims.UserId).
			Update("status", enums.IS_FRIEND).Error; err != nil {
			log.Logger.Error().Err(err).Msg("更新失败")
			return err
		}
		return nil
	})
}
