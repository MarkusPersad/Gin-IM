package handler

import (
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/response"
	"Gin-IM/pkg/token"
	"Gin-IM/pkg/validates"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AddFriend 添加好友
// @Summary 添加好友
// @Description 添加好友
// @Tags 好友
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param friend_request body request.FriendRequest true "friend_request"
// @Success 200 {object} response.Response{data=string} "成功"
// @Failure 200 {object} response.Response{data=string} "失败"
// @Router /api/friend/add [post]
func (h *Handlers) AddFriend(ctx *gin.Context) {
	claims, err := token.ExtractClaims(ctx)
	if err != nil {
		err = ctx.Error(err)
		return
	}
	if len(claims.UserId) == 0 {
		err = ctx.Error(exception.ErrTokenEmpty)
		return
	}
	if str := h.db.GetValue(ctx, defines.USER_TOKEN_KEY+claims.UserId); str == "" {
		err = ctx.Error(exception.ErrLoginTimeout)
		return
	}
	var friendRequest request.FriendRequest
	if err := ctx.BindJSON(&friendRequest); err != nil {
		err = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&friendRequest); err != nil {
		err = ctx.Error(err)
		return
	}
	if err := h.db.AddFriend(ctx, claims, friendRequest); err != nil {
		err = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "添加好友成功", nil))
}

// GetFriendList 获取好友列表
// @Summary 获取好友列表
// @Description 获取好友列表
// @Tags 好友
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/friend/list [get]
func (h *Handlers) GetFriendList(ctx *gin.Context) {
	claims, err := token.ExtractClaims(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if len(claims.UserId) == 0 {
		_ = ctx.Error(exception.ErrTokenEmpty)
		return
	}
	if str := h.db.GetValue(ctx, defines.USER_TOKEN_KEY+claims.UserId); str == "" {
		_ = ctx.Error(exception.ErrLoginTimeout)
		return
	}
	if friends, err := h.db.GetFriendList(ctx, claims); err != nil {
		_ = ctx.Error(err)
		return
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "获取好友列表成功", friends))
	}
}
