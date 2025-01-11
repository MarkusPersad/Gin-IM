package handler

import (
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/response"
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
		err = ctx.Error(exception.ErrCheckCode)
		return
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "获取验证码成功", database64))
	}
}

// Register 注册
// @Summary 注册
// @Description 注册
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
		err = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&register); err != nil {
		err = ctx.Error(err)
		return
	}
	capt := utils.NewCaptcha(h.db)
	if err := capt.Verify(register.CheckCodeKey, register.CheckCode, true); err != nil {
		err = ctx.Error(exception.ErrCheckCode)
		return
	}
	if err := h.db.Register(ctx, register); err != nil {
		err = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "注册成功", nil))
}
