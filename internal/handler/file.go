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

// UploadFile godoc
// @Summary 上传文件
// @Description 上传文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param fileUpload body request.FileUpload true "文件上传信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/file/upload [post]
func (h *Handlers) UploadFile(ctx *gin.Context) {
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
	var upload request.FileUpload
	if err := ctx.BindJSON(&upload); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(ctx, &upload); err != nil {
		_ = ctx.Error(err)
		return
	}
	if uploadUrl, err := h.db.UploadFile(ctx, claims, upload); err != nil {
		_ = ctx.Error(exception.ErrUploadFile)
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "上传成功", uploadUrl))
	}
}

// GetShortUrl godoc
// @Summary 获取文件下载地址
// @Description 获取文件下载地址
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param fileDownload body request.FileDownload true "文件下载信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败受到"
// @Router /api/file/download [post]
func (h *Handlers) GetShortUrl(ctx *gin.Context) {
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
	var fileDownload request.FileDownload
	if err := ctx.BindJSON(&fileDownload); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(ctx, &fileDownload); err != nil {
		_ = ctx.Error(err)
		return
	}
	if url := h.db.GetShortURL(ctx, claims, fileDownload); url == "" {
		_ = ctx.Error(exception.ErrFileUrl)
		return
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "获取成功", url))
	}
}

// DeleteFile godoc
// @Summary 删除文件
// @Description 删除文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param fileDelete body request.FileDelete true "文件删除信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/file/delete [post]
func (h *Handlers) DeleteFile(ctx *gin.Context) {
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
	var fileDeletes request.FileDeletes
	if err := ctx.BindJSON(&fileDeletes); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(ctx, &fileDeletes); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.DeleteFile(ctx, claims, fileDeletes.Deletes); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "删除成功", nil))
}

func (h *Handlers) MergeFile(ctx *gin.Context) {
	var merge request.FileMerge
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
	if err := ctx.BindJSON(&merge); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(ctx, &merge); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.MergeFile(ctx, claims, merge); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "合并成功", nil))
}

func (h *Handlers) GetTrash(ctx *gin.Context) {
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
	if recoveries := h.db.GetFileTrash(ctx, claims); len(recoveries) == 0 {
		ctx.JSON(http.StatusOK, response.Success(0, "获取成功", nil))
	} else {
		ctx.JSON(http.StatusOK, response.Success(0, "获取成功", recoveries))
	}
}

func (h *Handlers) RecoveryFile(ctx *gin.Context) {
	var recoveries request.FileRecoveryList
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
	if err := ctx.BindJSON(&recoveries); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(ctx, &recoveries); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.Recovery(ctx, claims, recoveries.Recoveries); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "恢复成功", nil))
}
