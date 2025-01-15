package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/types"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"io"
)

type FileService interface {
	UploadFile(ctx context.Context, claims *types.GIClaims, file io.Reader, objectName, fileName, md5, sha1 string, fileSize int64) error
	GetShortURL(ctx context.Context, download request.FileDownload) string
	DeleteFile(ctx context.Context, claims *types.GIClaims, deletes request.FileDelete) error
}

func (s *service) UploadFile(ctx context.Context, claims *types.GIClaims, file io.Reader, objectName, fileName, md5, sha1 string, fileSize int64) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		if s.CheckIsExist(ctx, md5, sha1) {
			return nil
		}
		if err := s.UploadFileToDB(ctx, claims, objectName, fileName, md5, sha1); err != nil {
			log.Logger.Error().Err(err).Msg("failed to create file")
			return exception.ErrUploadFile
		}
		if err := s.minClient.UploadFile(ctx, file, objectName, fileSize); err != nil {
			log.Logger.Error().Err(err).Str("objectName", objectName).Str("fileName", fileName).Str("md5", md5).Str("sha1", sha1).Msg("failed to upload file to MinIO")
			return exception.ErrUploadFile
		}
		return nil
	})
}
func (s *service) UploadFileToDB(ctx context.Context, claims *types.GIClaims, objectName, fileName, md5, sha1 string) error {
	var file model.File
	file.ObjectName = objectName
	file.FileName = fileName
	file.Owner = claims.UserId
	file.Md5 = md5
	file.Sha1 = sha1
	if err := s.GetDB(ctx).Model(&model.File{}).
		FirstOrCreate(&file, model.File{Md5: file.Md5, Sha1: file.Sha1, Owner: file.Owner}).
		Error; err != nil {
		log.Logger.Error().Err(err).Msg("failed to create file")
		return err
	}
	return nil
}
func (s *service) CheckIsExist(ctx context.Context, md5, sha1 string) bool {
	if err := s.GetDB(ctx).Model(&model.File{}).Where("md5 = ? and sha1 = ?", md5, sha1).First(&model.File{}).Error; err != nil {
		return false
	}
	return true
}

func (s *service) GetShortURL(ctx context.Context, download request.FileDownload) string {
	var shortURL string
	err := s.Transaction(ctx, func(ctx context.Context) error {
		if !s.CheckIsExist(ctx, download.Md5, download.Sha1) {
			return exception.ErrNotFound
		}
		objectName, fileName := s.GetObjectName(ctx, download)
		if objectName == "" || fileName == "" {
			return exception.ErrFileUrl
		}
		url, err := s.minClient.GetFileSign(ctx, objectName)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get file sign")
			return err
		}
		shortURL = url
		return nil
	})
	if err != nil {
		return ""
	}
	return shortURL
}
func (s *service) GetObjectName(ctx context.Context, download request.FileDownload) (objectName, fileName string) {
	var file model.File
	if err := s.GetDB(ctx).Model(&model.File{}).
		Where("md5 = ?", download.Md5).
		Where("sha1 = ?", download.Sha1).
		First(&file).Error; err != nil {
		log.Logger.Error().Err(err).Msg("failed to get file")
		return "", ""
	}
	return file.ObjectName, file.FileName
}

func (s *service) DeleteFile(ctx context.Context, claims *types.GIClaims, deletes request.FileDelete) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var objectNames []string
		if err := s.GetDB(ctx).Model(&model.File{}).
			Where("id IN ?", deletes.Seleted).
			Where("owner = ?", claims.UserId).
			Pluck("objectname", &objectNames).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Logger.Error().Err(err).Msg("failed to get file")
				return exception.ErrNotFound
			}
			// 如果没有找到文件记录，则直接返回
			return nil
		}
		if err := s.GetDB(ctx).Unscoped().
			Where("id IN ?", deletes.Seleted).
			Where("owner = ?", claims.UserId).
			Delete(&model.File{}).Error; err != nil {
			log.Logger.Error().Err(err).Msg("failed to delete file")
			return exception.ErrFileDelete
		}
		if err := s.minClient.DeleteFiles(ctx, true, objectNames...); err != nil {
			return exception.ErrFileDelete
		}
		return nil
	})
}
