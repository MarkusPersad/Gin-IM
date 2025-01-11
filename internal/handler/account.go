package handler

import (
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/response"
	"Gin-IM/pkg/utils"
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
		ctx.Error(exception.ErrCheckCode)
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "获取验证码成功", database64))
	}
}
