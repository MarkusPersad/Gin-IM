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
		return
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
// @Param fileDeletes body request.FileDeletes true "文件删除信息"
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

// MergeFile godoc
// @Summary 合并文件
// @Description 合并文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param merge body request.FileMerge true "文件合并信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/file/merge [post]
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

// GetTrash godoc
// @Summary 获取回收站文件列表
// @Description 获取回收站文件列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/file/trash [get]
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
	recoveries := h.db.GetFileTrash(ctx, claims)
	ctx.JSON(http.StatusOK, response.Success(0, "获取成功", recoveries))
}

// RecoveryFile godoc
// @Summary 恢复文件
// @Description 恢复文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param recoveries body request.FileRecoveryList true "文件恢复信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/file/recovery [post]
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

// PushPartsInfo godoc
// @Summary 上传分片信息
// @Description 上传分片信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param partInfo body request.PartInfo true "分片信息"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败"
// @Router /api/file/pushparts [post]
func (h *Handlers) PushPartsInfo(ctx *gin.Context) {
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
	var parts request.PartInfo
	if err := ctx.BindJSON(&parts); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(ctx, &parts); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.PushPartsInfo(ctx, parts); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "分片信息上传成功", nil))
}
