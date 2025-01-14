package handler

import (
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/response"
	"Gin-IM/pkg/token"
	"Gin-IM/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"path/filepath"
	"strings"
)

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
		_ = ctx.Error(exception.ErrUploadFile)
		return
	}
	md5Value := ctx.PostForm("md5")
	if md5Value == "" {
		log.Logger.Error().Msg("md5 is empty")
		_ = ctx.Error(exception.ErrUploadFile)
		return
	}
	sha1Value := ctx.PostForm("sha1")
	if sha1Value == "" {
		log.Logger.Error().Msg("sha1 is empty")
		_ = ctx.Error(exception.ErrUploadFile)
		return
	}
	fileType := utils.GetFileType(strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."))
	objectName := fileType + "/" + md5Value + sha1Value + filepath.Ext(fileHeader.Filename)
	if h.db.CheckIsExist(ctx, md5Value, sha1Value) {
		ctx.JSON(http.StatusOK, response.Success(0, "文件已存在", nil))
		return
	}
	if _, err := h.minioClient.UploadFile(ctx, file, "chat", objectName, fileHeader.Size); err != nil {
		log.Logger.Error().Err(err).Msg("failed to upload file")
		_ = ctx.Error(exception.ErrUploadFile)
		return
	}
	if err := h.db.UploadFile(ctx, claims, fileHeader.Filename, utils.GetFileTypeEnum(fileType), md5Value, sha1Value); err != nil {
		log.Logger.Error().Err(err).Msg("failed to upload file")
		_ = ctx.Error(exception.ErrUploadFile)
		return
	}
	ctx.JSON(http.StatusOK, response.Success(0, "上传成功", nil))
}
