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
	"github.com/rs/zerolog/log"
	"net/http"
	"path/filepath"
	"strings"
)

// UploadFile godoc
// @Summary 上传文件
// @Description 用户上传文件到服务器，并存储在 MinIO 中
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer Token令牌"
// @Param file formData file true "要上传的文件"
// @Param md5 formData string true "文件的 MD5 值"
// @Param sha1 formData string true "文件的 SHA1 值"
// @Success 200 {object} response.Response "成功"
// @Failure 200 {object} response.Response "失败受到"
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
	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to get file")
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	md5Value := ctx.PostForm("md5")
	if md5Value == "" {
		log.Logger.Error().Msg("md5 is empty")
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	sha1Value := ctx.PostForm("sha1")
	if sha1Value == "" {
		log.Logger.Error().Msg("sha1 is empty")
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	fileType := utils.GetFileType(strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."))
	objectName := fileType + "/" + md5Value + sha1Value + filepath.Ext(fileHeader.Filename)
	if err := h.db.UploadFile(ctx, claims, file, objectName, fileHeader.Filename, md5Value, sha1Value, fileHeader.Size); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "上传成功", fileHeader.Filename))
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
	if err := validates.Validate(&fileDownload); err != nil {
		_ = ctx.Error(err)
		return
	}
	if url := h.db.GetShortURL(ctx, fileDownload); url == "" {
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
func (h Handlers) DeleteFile(ctx *gin.Context) {
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
	var fileDelete request.FileDelete
	if err := ctx.BindJSON(&fileDelete); err != nil {
		_ = ctx.Error(exception.ErrBadRequest)
		return
	}
	if err := validates.Validate(&fileDelete); err != nil {
		_ = ctx.Error(err)
		return
	}
	if err := h.db.DeleteFile(ctx, claims, fileDelete); err != nil {
		_ = ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "删除成功", nil))
}
