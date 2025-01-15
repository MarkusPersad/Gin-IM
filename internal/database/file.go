package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/types"
	"context"
	"github.com/rs/zerolog/log"
)

type FileService interface {
	UploadFile(ctx context.Context, claims *types.GIClaims, objectName string, fileType int8, md5, sha1 string) error
	CheckIsExist(ctx context.Context, md5, sha1 string) bool
	GetShortUrl(ctx context.Context, request request.FileDownload) (string, string)
}

func (s *service) UploadFile(ctx context.Context, claims *types.GIClaims, objectName string, fileType int8, md5, sha1 string) error {
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

func (s *service) CheckIsExist(ctx context.Context, md5, sha1 string) bool {
	if err := s.GetDB(ctx).Model(&model.File{}).Where("md5 = ? and sha1 = ?", md5, sha1).First(&model.File{}).Error; err != nil {
		return false
	}
	return true
}

func (s *service) GetShortUrl(ctx context.Context, request request.FileDownload) (string, string) {
	var file model.File
	err := s.Transaction(ctx, func(ctx context.Context) error {
		if err := s.GetDB(ctx).Model(&model.File{}).
			Where("md5 = ?", request.Md5).
			Where("sha1 = ?", request.Sha1).
			First(&file).Error; err != nil {
			log.Logger.Error().Err(err).Msg("failed to get file")
			return err
		}
		return nil
	})
	if err != nil {
		return "", ""
	}
	return file.Md5 + file.Sha1, file.ObjectName
}
