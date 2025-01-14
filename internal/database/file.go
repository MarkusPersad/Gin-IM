package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/types"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type FileService interface {
	UploadFile(ctx *gin.Context, claims *types.GIClaims, objectName string, fileType int8, md5, sha1 string) error
	CheckIsExist(ctx *gin.Context, md5, sha1 string) bool
}

func (s *service) UploadFile(ctx *gin.Context, claims *types.GIClaims, objectName string, fileType int8, md5, sha1 string) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var file model.File
		file.FileType = fileType
		file.ObjectName = objectName
		file.Owner = claims.UserId
		file.Md5 = md5
		file.Sha1 = sha1
		if err := s.GetDB(ctx).Model(&model.File{}).Create(&file).Error; err != nil {
			log.Logger.Error().Err(err).Msg("failed to create file")
			return err
		}
		return nil
	})
}

func (s *service) CheckIsExist(ctx *gin.Context, md5, sha1 string) bool {
	if err := s.GetDB(ctx).Model(&model.File{}).Where("md5 = ? and sha1 = ?", md5, sha1).First(&model.File{}).Error; err != nil {
		return false
	}
	return true
}
