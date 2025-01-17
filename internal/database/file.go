package database

import (
	"Gin-IM/internal/model"
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/enums"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/request"
	"Gin-IM/pkg/types"
	"Gin-IM/pkg/utils"
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"path/filepath"
	"strings"
)

type FileService interface {
	UploadFile(ctx context.Context, claims *types.GIClaims, upload request.FileUpload) (*types.UploadUrls, error)
	GetShortURL(ctx context.Context, claims *types.GIClaims, download request.FileDownload) string
	DeleteFile(ctx context.Context, claims *types.GIClaims, deletes []request.FileDelete) error
	MergeFile(ctx context.Context, claims *types.GIClaims, merge request.FileMerge) error
	GetFileTrash(ctx context.Context, claims *types.GIClaims) []request.FileRecovery
	Recovery(ctx context.Context, claims *types.GIClaims, recoveries []request.FileRecovery) error
	PushPartsInfo(ctx context.Context, parts request.PartInfo) error
}

// UploadFile 上传文件服务。该函数处理文件上传的请求，根据文件信息生成对象名，
// 并通过事务处理方式初始化上传过程，包括检查文件是否存在、获取上传ID、生成上传URL等。
// 参数:
//
//	ctx - 上下文，用于传递请求范围的数据和控制超时。
//	claims - 用户声明，包含用户相关信息。
//	upload - 文件上传请求，包含文件信息。
//
// 返回值:
//
//	*types.UploadUrls - 包含上传URL和上传ID的信息，如果上传初始化成功。
//	error - 错误信息，如果上传初始化过程中发生错误。
func (s *service) UploadFile(ctx context.Context, claims *types.GIClaims, upload request.FileUpload) (*types.UploadUrls, error) {
	var uploadUrls *types.UploadUrls

	// 使用事务处理上传过程，确保数据一致性。
	err := s.Transaction(ctx, func(ctx context.Context) error {
		// 构造对象名，基于文件类型、MD5、SHA1和文件扩展名。
		objectName := utils.GetFileType(filepath.Ext(upload.FileName)) + "/" + upload.Md5 + upload.Sha1 + filepath.Ext(upload.FileName)

		// 初始化上传信息。
		uploadInfo := types.UploadInfo{
			ObjectName:  objectName,
			ChunkSize:   upload.ChunkSize,
			ChunkNumber: upload.ChunkNumber,
		}

		// 从缓存中获取上传ID和已完成的分片列表。
		uploadInfo.UploadId = s.GetValue(ctx, defines.UPLOAD_ID+objectName)
		uploadInfo.Completed = s.GetList(ctx, defines.COMPLETED_PARTS+uploadInfo.UploadId)

		// 检查文件是否已存在，如果存在，则更新数据库并结束上传初始化过程。
		if file, isExist := s.CheckIsExist(ctx, upload.Md5, upload.Sha1); isExist {
			if err := s.UploadFileToDB(ctx, claims, file.UploadId, objectName, file.FileName, file.Md5, file.Sha1, enums.FILEUPLOADED); err != nil {
				return err
			} else {
				return nil
			}
		}

		// 检查MinIO中是否已存在该对象，如果存在，则更新数据库并结束上传初始化过程。
		if _, err := s.minClient.StatusObject(ctx, objectName); err == nil {
			if err := s.UploadFileToDB(ctx, claims, uploadInfo.UploadId, objectName, upload.FileName, upload.Md5, upload.Sha1, enums.FILEUPLOADED); err != nil {
				return err
			} else {
				return nil
			}
		}

		// 初始化分片上传，获取上传URL。
		uploadUrl, err := s.minClient.InitMutiparts(ctx, uploadInfo)
		if err != nil {
			return err
		} else {
			// 设置上传ID到缓存，以便后续使用。
			if s.SetAndTime(ctx, defines.UPLOAD_ID+objectName, uploadUrl.UploadId, defines.FILE_SHORT_SIGN*60*60) != nil {
				log.Logger.Error().Err(err).Msg("设置上传ID失败")
				return err
			}
		}

		// 更新数据库，记录上传开始状态。
		if err := s.UploadFileToDB(ctx, claims, uploadUrl.UploadId, objectName, upload.FileName, upload.Md5, upload.Sha1, enums.FILEUPLOADING); err != nil {
			return err
		}

		// 保存上传URL，供外部使用。
		uploadUrls = uploadUrl
		return nil
	})

	// 如果事务处理过程中发生错误，返回错误信息。
	if err != nil {
		return nil, err
	}

	// 返回上传URL信息。
	return uploadUrls, nil
}

// UploadFileToDB 将文件信息上传到数据库
// 该函数接收上传文件的元数据，包括上传ID、对象名称、文件名以及文件的MD5和SHA1哈希值，
// 并将这些信息与用户ID一起存储在数据库中
func (s *service) UploadFileToDB(ctx context.Context, claims *types.GIClaims, uploadId, objectName, fileName, md5, sha1 string, status enums.FileEnum) error {
	// 初始化一个File结构体实例，用于存储文件信息
	var file model.File

	// 为文件结构体赋值
	file.UploadId = uploadId
	file.ObjectName = objectName
	file.FileName = fileName
	// 从claims中获取用户ID，作为文件的拥有者
	file.Owner = claims.UserId
	file.Md5 = md5
	file.Sha1 = sha1
	file.Status = int8(status)
	// 使用数据库模型将文件信息保存到数据库中
	// 如果数据库中已存在相同MD5、SHA1和拥有者的文件记录，则不会创建新记录
	if err := s.GetDB(ctx).Model(&model.File{}).
		FirstOrCreate(&file, model.File{Md5: file.Md5, Sha1: file.Sha1, Owner: file.Owner}).
		Error; err != nil {
		// 如果创建文件记录失败，记录错误日志并返回错误
		log.Logger.Error().Err(err).Msg("failed to create file")
		return err
	}

	// 如果文件记录成功创建或已存在，则不返回任何错误
	return nil
}

// CheckIsExist 检查具有指定MD5和SHA1哈希值的文件是否存在。
// 该方法接收一个服务实例的指针，一个上下文，以及文件的MD5和SHA1哈希值作为参数。
// 它返回一个指向找到的文件对象的指针和一个布尔值，指示是否成功找到文件。
// 如果没有找到匹配的文件或者数据库查询发生错误，则返回nil和false。
func (s *service) CheckIsExist(ctx context.Context, md5, sha1 string) (*model.File, bool) {
	var file model.File
	// 使用MD5、SHA1哈希值和上传状态来查询数据库中的文件。
	// 如果查询到文件，则返回该文件对象和true。
	// 如果未找到文件或发生错误，则返回nil和false。
	if err := s.GetDB(ctx).Model(&model.File{}).
		Where("md5 = ?", md5).
		Where("sha1 = ?", sha1).
		Where("status = ?", enums.FILEUPLOADED).
		First(&file).Error; err != nil {
		return nil, false
	}
	return &file, true
}

// GetShortURL 生成文件的短链接URL。
// 该方法首先检查文件是否存在，然后获取文件的签名URL，并将其作为短链接返回。
// 参数:
//
//	ctx - 上下文，用于传递请求范围的数据和控制超时。
//	claims - 包含用户信息的令牌声明。
//	download - 包含文件下载信息的请求对象。
//
// 返回值:
//
//	文件的短链接URL字符串，如果生成失败则返回空字符串。
func (s *service) GetShortURL(ctx context.Context, claims *types.GIClaims, download request.FileDownload) string {
	var shortURL string
	// 使用事务确保操作的原子性。
	err := s.Transaction(ctx, func(ctx context.Context) error {
		// 检查文件是否存在，如果不存在则返回错误。
		if _, isExist := s.CheckIsExist(ctx, download.Md5, download.Sha1); !isExist {
			return exception.ErrNotFound
		}
		// 获取文件的对象名称，如果获取失败则返回错误。
		objectName := s.GetObjectName(ctx, claims, download)
		if objectName == "" {
			return exception.ErrNotFound
		}
		// 获取文件的签名URL，如果获取失败则记录错误并返回。
		url, err := s.minClient.GetFileSign(ctx, objectName)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get file sign")
			return err
		}
		// 将获取到的URL保存到变量中，以便后续返回。
		shortURL = url
		return nil
	})
	// 如果事务执行过程中出现错误，则返回空字符串。
	if err != nil {
		return ""
	}
	// 返回生成的短链接URL。
	return shortURL
}

// GetObjectName 根据文件的MD5、SHA1值、状态和所有者获取文件的名称。
// 该方法主要用于在给定文件下载请求和用户信息的条件下，从数据库中查询并返回文件的名称。
// 如果找不到符合条件的文件，该方法将返回一个空字符串。
//
// 参数:
//
//	ctx - 上下文，用于传递请求范围的信息，如数据库会话。
//	claims - 用户声明，包含用户信息，如用户ID，用于确定文件的所有者。
//	download - 文件下载请求，包含文件的MD5和SHA1值，用于定位文件。
//
// 返回值:
//
//	返回找到的文件的名称，如果未找到则返回空字符串。
func (s *service) GetObjectName(ctx context.Context, claims *types.GIClaims, download request.FileDownload) string {
	var file model.File
	// 使用MD5、SHA1值、文件上传状态和用户ID查询数据库中的文件记录。
	if err := s.GetDB(ctx).Model(&model.File{}).
		Where("md5 = ?", download.Md5).
		Where("sha1 = ?", download.Sha1).
		Where("status = ?", enums.FILEUPLOADED).
		Where("owner = ?", claims.UserId).
		First(&file).Error; err != nil {
		// 如果查询过程中出现错误，记录错误信息并返回空字符串。
		log.Logger.Error().Err(err).Msg("failed to get file")
		return ""
	}
	// 返回查询到的文件的名称。
	return file.ObjectName
}

// DeleteFile 删除文件服务。
// 该函数接收一个上下文、用户声明和一个文件删除请求。
// 它通过事务执行删除操作，确保数据一致性。
func (s *service) DeleteFile(ctx context.Context, claims *types.GIClaims, deletes []request.FileDelete) error {
	// 使用事务确保删除操作的原子性。
	return s.Transaction(ctx, func(ctx context.Context) error {
		var objectNames []string
		// 遍历用户选择的文件，获取它们的对象名称。
		for _, file := range deletes {
			// 根据文件的MD5、SHA1和文件名获取对象名称。
			if objectName := s.GetObjectName(ctx, claims, request.FileDownload{
				Md5:      file.Md5,
				Sha1:     file.Sha1,
				FileName: file.FileName,
			}); objectName != "" {
				// 如果对象名称不为空，添加到待删除列表中。
				objectNames = append(objectNames, objectName)
			} else {
				// 如果对象名称为空，表示文件未找到，返回错误。
				return exception.ErrNotFound
			}
		}
		// 使用获取的对象名称列表执行文件的批量删除操作。
		if err := s.GetDB(ctx).
			Where("objectname IN ?", objectNames).
			Delete(&model.File{}).Error; err != nil {
			// 如果删除失败，记录错误日志并返回自定义错误。
			log.Logger.Error().Err(err).Msg("failed to delete file")
			return exception.ErrFileDelete
		}
		return nil
	})
}

// MergeFile 合并文件片段成一个完整的文件。
// 该函数通过给定的文件合并请求，检查并合并文件片段，确保文件的完整性和正确性。
func (s *service) MergeFile(ctx context.Context, claims *types.GIClaims, merge request.FileMerge) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var uploadId string
		// 构造对象名称，确保文件类型、MD5、SHA1和文件扩展名的唯一组合。
		objectName := utils.GetFileType(filepath.Ext(merge.FileName)) + "/" + merge.Md5 + merge.Sha1 + filepath.Ext(merge.FileName)

		// 尝试从上下文中获取上传ID，如果存在则使用，否则从数据库中查询。
		if uploadid := s.GetValue(ctx, defines.UPLOAD_ID+objectName); uploadid != "" {
			uploadId = uploadid
		} else {
			// 查询数据库以获取上传ID。
			if err := s.GetDB(ctx).Model(&model.File{}).
				Where("objectname = ?", objectName).
				Where("owner = ?", claims.UserId).
				Where("status = ?", enums.FILEUPLOADING).
				Pluck("uploadid", &uploadId).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Logger.Error().Err(err).Msg("failed to get file")
					return exception.ErrNotFound
				}
			}
		}

		// 更新数据库中的文件状态为已上传。
		if err := s.GetDB(ctx).Model(&model.File{}).
			Where("md5 = ?", merge.Md5).
			Where("sha1 = ?", merge.Sha1).
			Where("owner = ?", claims.UserId).
			Update("status", enums.FILEUPLOADED).Error; err != nil {
			log.Logger.Error().Err(err).Msg("failed to update file")
			return exception.ErrFileUploading
		}

		// 调用MinIO客户端合并文件片段。
		if err := s.minClient.MergeMutipartsUpload(ctx, uploadId, objectName); err != nil {
			return exception.ErrFileUploading
		}

		// 检查合并后的文件状态。
		if objectInfo, err := s.minClient.StatusObject(ctx, objectName); err != nil {
			log.Logger.Error().Err(err).Msg("failed to get object info")
			return err
		} else {
			// 校验文件的SHA1值以确保文件完整性。
			if merge.Sha1 != objectInfo.ChecksumSHA1 {
				_ = s.minClient.DeleteFiles(ctx, true, objectName)
				return exception.ErrFileUploading
			}
		}

		// 删除缓存中的上传ID。
		if err := s.DelValue(ctx, defines.UPLOAD_ID+objectName, defines.COMPLETED_PARTS+uploadId); err != nil {
			log.Logger.Error().Err(err).Msg("failed to del value")
			return err
		}
		return nil
	})
}

func (s *service) GetFileTrash(ctx context.Context, claims *types.GIClaims) []request.FileRecovery {
	var files []request.FileRecovery
	err := s.Transaction(ctx, func(ctx context.Context) error {
		if err := s.GetDB(ctx).Unscoped().Model(&model.File{}).
			Select("md5,sha1,filename").
			Where("owner = ?", claims.UserId).
			Where("status = ?", enums.FILEUPLOADED).
			Where("deleted_at IS NOT NULL").
			Scan(&files).Error; err != nil {
			log.Logger.Error().Err(err).Msg("failed to get file")
			return exception.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return files
}

func (s *service) Recovery(ctx context.Context, claims *types.GIClaims, recoveries []request.FileRecovery) error {
	return s.Transaction(ctx, func(ctx context.Context) error {
		var objectNames []string
		for _, recovery := range recoveries {
			objectName := utils.GetFileType(filepath.Ext(recovery.FileName)) + "/" + recovery.Md5 + recovery.Sha1 + filepath.Ext(recovery.FileName)
			objectNames = append(objectNames, objectName)
		}
		if err := s.GetDB(ctx).Unscoped().Model(&model.File{}).
			Where("objectname IN ?", objectNames).
			Where("owner = ?", claims.UserId).
			Where("deleted_at IS NOT NULL").
			Update("deleted_at", nil).Error; err != nil {
			log.Logger.Error().Err(err).Msg("failed to recovery file")
			return exception.ErrFileRecovery
		}
		return nil
	})
}

// PushPartsInfo 接收并处理分片上传的信息。
// 该方法根据提供的parts信息，将分片编号存储，并设置其过期时间。
// 主要用于跟踪和管理大文件上传过程中的分片信息。
// 参数:
//
//	ctx - 上下文，用于传递请求范围的 deadline、取消信号等。
//	parts - 包含分片上传编号和其他信息的请求对象。
//
// 返回值:
//
//	如果操作成功，则返回nil；如果操作失败，则返回错误。
func (s *service) PushPartsInfo(ctx context.Context, parts request.PartInfo) error {
	// 将传入的分片编号字符串按逗号分割，转换为字符串切片。
	uploadParts := strings.Split(parts.PartNums, ",")

	// 调用SetListAndTime方法，设置已完成分片的列表和过期时间。
	// 如果操作失败，则记录错误日志并返回错误。
	if err := s.SetListAndTime(ctx, defines.COMPLETED_PARTS+parts.UploadId, uploadParts, 10*60); err != nil {
		log.Logger.Error().Err(err).Msg("failed to push parts info")
		return err
	}

	// 如果一切顺利，返回nil表示操作成功。
	return nil
}
