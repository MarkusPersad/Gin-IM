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
	//TODO 发送好友请求
	ctx.JSON(http.StatusOK, response.Success(0, "发送好友请求", nil))
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

// AddToBlackList 拉黑好友
// @Summary 拉黑好友
// @Description 拉黑好友
// @Tags 好友
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param friend_request body request.FriendRequest true "好友信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/friend/black [post]
func (h *Handlers) AddToBlackList(ctx *gin.Context) {
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
	var friendRequest request.FriendRequest
	if err := ctx.BindJSON(&friendRequest); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.AddToBlackList(ctx, claims, friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "拉黑成功", nil))
}

// GetBlackList 查询拉黑列表
// @Summary 查询拉黑列表
// @Description 查询拉黑列表
// @Tags 好友
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/friend/blacklist [get]
func (h *Handlers) GetBlackList(ctx *gin.Context) {
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
	if blackList, err := h.db.GetBlackList(ctx, claims); err != nil {
		_ = ctx.Error(err)
		return
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "查询成功", blackList))
	}
}

// CancelBlack 取消拉黑
// @Summary 取消拉黑
// @Description 取消拉黑
// @Tags 好友
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param friend_request body request.FriendRequest true "好友信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/friend/cancelblack [post]
func (h *Handlers) CancelBlack(ctx *gin.Context) {
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
	var friendRequest request.FriendRequest
	if err := ctx.BindJSON(&friendRequest); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.CancelBlack(ctx, claims, friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "取消拉黑成功", nil))
}

// DeleteFriend 删除好友
// @Summary 删除好友
// @Description 删除好友
// @Tags 好友
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param friend_request body request.FriendRequest true "好友信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/friend/delete [post]
func (h *Handlers) DeleteFriend(ctx *gin.Context) {
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
	var friendRequest request.FriendRequest
	if err := ctx.BindJSON(&friendRequest); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.DeleteFriend(ctx, claims, friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "删除好友成功", nil))
}

// AgreeFriendRequest 同意好友请求
// @Summary 同意好友请求
// @Description 同意好友请求
// @Tags 好友
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param friend_request body request.FriendRequest true "好友信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/friend/agree [post]
func (h *Handlers) AgreeFriendRequest(ctx *gin.Context) {
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
	var friendRequest request.FriendRequest
	if err := ctx.BindJSON(&friendRequest); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.AgreeFriendRequest(ctx, claims, friendRequest); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "同意好友请求成功", nil))
}
