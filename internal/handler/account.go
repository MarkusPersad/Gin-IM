package handler

import (
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/response"
	"Gin-IM/pkg/token"
	"Gin-IM/pkg/utils"
	"Gin-IM/pkg/validates"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetCaptcha 获取验证码
// @Summary 获取验证码
// @Description 获取验证码
// @Tags 账户管理
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response "返回结果"
// @Failure 200 {object} response.Response "返回结果"
// @Router /api/account/getcaptcha [get]
func (h *Handlers) GetCaptcha(ctx *gin.Context) {
	capt := utils.NewCaptcha(h.db)
	if database64, err := capt.Generate(); err != nil {
		_ = ctx.Error(exception.ErrCheckCode)
		return
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "获取验证码成功", database64))
	}
}

// Register 注册
// @Summary 注册
// @Description 处理用户注册请求
// @Tags 账户管理
// @Accept  json
// @Produce json
// @Param register body request.Register true "注册信息"
// @Success 200 {object} response.Response "返回结果"
// @Failure 200 {object} response.Response "返回结果"
// @Router /api/account/register [post]
func (h *Handlers) Register(ctx *gin.Context) {
	var register request.Register
	if err := ctx.BindJSON(&register); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&register); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := utils.NewCaptcha(h.db).Verify(register.CheckCodeKey, register.CheckCode, true); err != nil {
		_ = ctx.Error(exception.ErrCheckCode)
		return
	}
	if err := h.db.Register(ctx, register); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "注册成功", nil))
}

// Login 处理用户登录请求。
// @Summary 登陆
// @Description 处理用户登录请求。
// @Tags 账户管理
// @Accept  json
// @Produce  json
// @Param login body request.Login true "登录信息"
// @Success 200 {object} response.Response "返回结果"
// @Failure 200 {object} response.Response "返回结果"
// @Router /api/account/login [post]
func (h *Handlers) Login(ctx *gin.Context) {
	// 尝试将请求体中的JSON数据绑定到login变量。
	var login request.Login
	if err := ctx.BindJSON(&login); err != nil {
		// 如果绑定失败，返回错误响应并结束函数执行。
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}

	// 验证登录信息的合法性。
	if err := validates.Validate(&login); err != nil {
		// 如果验证失败，返回错误信息并结束函数执行。
		_ = ctx.Error(err)
		return
	}

	// 验证用户提供的验证码。
	if err := utils.NewCaptcha(h.db).Verify(login.CheckCodeKey, login.CheckCode, true); err != nil {
		// 如果验证码验证失败，返回错误信息并结束函数执行。
		_ = ctx.Error(exception.ErrCheckCode)
		return
	}

	// 调用数据库接口进行用户登录验证。
	if tokens, err := h.db.Login(ctx, login); err != nil {
		// 如果登录验证失败，返回错误信息并结束函数执行。
		_ = ctx.Error(err)
		return
	} else {
		// 如果登录成功，返回成功响应和生成的用户令牌。
		ctx.JSON(http.StatusOK, response.Success(0, "登录成功", tokens))
	}
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 获取用户信息
// @Tags 账户管理
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token令牌"
// @Success 200 {object} response.Response "返回结果"
// @Failure 200 {object} response.Response "返回结果"
// @Router /api/account/getuserinfo [get]
func (h *Handlers) GetUserInfo(ctx *gin.Context) {
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
	if user, err := h.db.GetUserInfo(ctx, claims); err != nil {
		_ = ctx.Error(err)
		return
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "获取用户信息成功", user))
	}
}

// Logout 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags 账户管理
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token令牌"
// @Success 200 {object} response.Response "返回结果"
// @Failure 200 {object} response.Response "返回结果"
// @Router /api/account/logout [get]
func (h *Handlers) Logout(ctx *gin.Context) {
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
	if err := h.db.Logout(ctx, claims); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "退出成功", nil))
}

// Search 搜索用户
// @Summary 搜索用户
// @Description 搜索用户
// @Tags 账户管理
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token令牌"
// @Param userSearch body request.UserSearch true "搜索信息"
// @Success 200 {object} response.Response "返回结果"
// @Failure 200 {object} response.Response "返回结果"
// @Router /api/account/search [post]
func (h *Handlers) Search(ctx *gin.Context) {
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
	var userSearch request.UserSearch
	if err := ctx.BindJSON(&userSearch); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&userSearch); err != nil {
		_ = ctx.Error(err)
		return
	}
	if user, err := h.db.Search(ctx, userSearch); err != nil {
		_ = ctx.Error(err)
		return
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "搜索成功", user))
	}
}
