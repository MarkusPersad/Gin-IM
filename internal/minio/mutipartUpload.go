package minio

import (
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/types"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
	"net/url"
	"strconv"
	"time"
)

// InitMutiparts 初始化多部分上传
// 该函数负责为给定的上传信息生成上传URLs
// 参数:
//
//	ctx - 上下文，用于传递请求范围的 deadline、取消信号等
//	upload - 包含上传信息的结构体，如桶名、对象名、分块数量等
//
// 返回值:
//
//	*types.UploadUrls - 包含上传URLs和上传ID的结构体指针
//	error - 如果操作失败，返回错误
func (s *MinIOStore) InitMutiparts(ctx context.Context, upload types.UploadInfo) (*types.UploadUrls, error) {
	// 初始化上传URLs结构体
	var uploadUrls = &types.UploadUrls{}

	// 确保桶存在，如果不存在则创建
	if err := s.CreateBucket(ctx, bucket, ""); err != nil {
		log.Logger.Error().Err(err).Msg("failed to create bucket")
		return nil, err
	}

	// 验证分块大小是否符合要求，分块数量和分块大小需要满足特定条件
	if upload.ChunkNumber <= 0 || upload.ChunkSize <= defines.MIN_CHUNK_SIZE {
		log.Logger.Error().Msg("ChunkSize Must greater 5 MB")
		return nil, exception.ErrBadRequest
	}

	// 处理单块上传的情况
	if upload.ChunkNumber == 1 {
		if presignedURL, err := s.PresignedPutObject(ctx, bucket, upload.ObjectName, defines.FILE_SHORT_SIGN*time.Hour); err != nil {
			log.Logger.Error().Err(err).Msg("failed to get file sign")
			return nil, err
		} else {
			uploadUrls.UploadId = defines.SINGLE_UPLOAD_ID
			uploadUrls.Urls = append(uploadUrls.Urls, presignedURL.String())
			return uploadUrls, nil
		}
	}

	// 处理多块上传的情况，首先获取或生成上传ID
	if upload.UploadId == "" {
		uploadId, err := s.NewMultipartUpload(ctx, bucket, upload.ObjectName, minio.PutObjectOptions{
			PartSize: upload.ChunkSize,
		})
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to new multipart upload")
			return nil, err
		}
		uploadUrls.UploadId = uploadId
	} else {
		uploadUrls.UploadId = upload.UploadId
	}

	// 获取已上传的分块信息，并标记为已完成
	if len(upload.Completed) == 0 {
		if objectParts, err := s.getListParts(ctx, upload.ObjectName, uploadUrls.UploadId); err == nil && len(objectParts) != 0 {
			for _, part := range objectParts {
				uploadUrls.Completed = append(uploadUrls.Completed, strconv.Itoa(part.PartNumber))
			}
		}
	} else {
		uploadUrls.Completed = upload.Completed
	}
	// 为每个分块生成预签名URL
	urlValues := url.Values{}
	urlValues.Set("uploadId", uploadUrls.UploadId)
	for num := 1; num <= upload.ChunkNumber; num++ {
		urlValues.Set("partNumber", strconv.Itoa(num))
		if presignedURL, err := s.Presign(ctx, "PUT", bucket, upload.ObjectName, defines.FILE_SHORT_SIGN*time.Hour, urlValues); err != nil {
			log.Logger.Error().Err(err).Msg("failed to get file sign")
			return nil, err
		} else {
			uploadUrls.Urls = append(uploadUrls.Urls, presignedURL.String())
		}
	}

	// 返回包含所有上传URLs和上传ID的结构体
	return uploadUrls, nil
}

// getListParts 获取对象的分片编号列表。
// 该方法通过循环调用 ListObjectParts 方法来获取对象的所有分片信息，直到所有分片都被列出（isTruncated 为 false）。
// 参数:
//   - ctx: 上下文，用于传递请求的上下文信息。
//   - objectName: 对象名称，标识要列出分片的对象。
//   - uploadId: 上传 ID，标识特定的上传操作。
//
// 返回值:
//   - []int: 包含所有分片编号的切片。
//   - error: 如果列出分片的过程中发生错误，则返回该错误。
func (s *MinIOStore) getListParts(ctx context.Context, objectName, uploadId string) ([]minio.ObjectPart, error) {
	// 初始化分片切片。
	var parts []minio.ObjectPart
	// 初始化起始编号为 0，用于标记从哪个分片开始列出。
	startNum := 0
	// 初始化 isTruncated 为 true，以确保至少进行一次列出操作。
	isTruncated := true

	// 循环列出分片，直到所有分片都被列出。
	for isTruncated {
		// 调用 ListObjectParts 方法列出分片。
		if partsResult, err := s.ListObjectParts(ctx, bucket, objectName, uploadId, startNum, defines.CHUNK_NUM); err != nil {
			// 如果发生错误，记录错误日志并返回错误。
			log.Logger.Error().Err(err).Msg("failed to list parts")
			return nil, err
		} else {
			// 更新是否还有更多分片的标志。
			isTruncated = partsResult.IsTruncated
			// 更新下一次列出操作的起始编号。
			startNum = partsResult.NextPartNumberMarker
			// 遍历当前列出的分片，并将分片添加到切片中。
			for _, part := range partsResult.ObjectParts {
				parts = append(parts, part)
			}
		}
	}

	// 返回包含所有分片编号的切片和 nil 错误，表示操作成功。
	return parts, nil
}

// MergeMutipartsUpload 合并分片上传的各个部分。
// 该方法首先获取所有已上传的分片信息，然后将这些分片信息合并，
// 以完成整个对象的上传过程。这一过程涉及到两个主要步骤：
// 1. 调用getListParts方法获取所有已上传的分片信息。
// 2. 调用CompleteMultipartUpload方法完成分片上传的合并。
// 参数:
//   - ctx: 上下文，用于传递请求相关的配置和元数据。
//   - uploadId: 分片上传的唯一标识符。
//   - objectName: 要上传的对象的名称。
//
// 返回值:
//   - error: 如果在获取分片列表或完成分片上传过程中遇到错误，则返回该错误。
func (s *MinIOStore) MergeMutipartsUpload(ctx context.Context, uploadId, objectName string) error {
	// 初始化一个空的completeParts切片，用于存储所有分片的完成信息。
	var completeParts []minio.CompletePart

	// 尝试获取分片列表。
	if parts, err := s.getListParts(ctx, objectName, uploadId); err != nil {
		// 如果获取分片列表时发生错误，记录错误日志并返回错误。
		log.Logger.Error().Err(err).Msg("failed to get list parts")
		return err
	} else {
		// 遍历获取到的分片列表，将每个分片的信息添加到completeParts中。
		for _, part := range parts {
			completeParts = append(completeParts, minio.CompletePart{
				PartNumber: part.PartNumber,
				ETag:       part.ETag,
			})
		}
	}

	// 尝试完成分片上传的合并。
	if _, err := s.CompleteMultipartUpload(ctx, bucket, objectName, uploadId, completeParts, minio.PutObjectOptions{}); err != nil {
		// 如果完成分片上传时发生错误，记录错误日志并返回错误。
		log.Logger.Error().Err(err).Msg("failed to complete multipart upload")
		return err
	}

	// 如果一切顺利，返回nil表示操作成功。
	return nil
}

// StatusObject 获取指定对象的状态信息。
//
// 该方法主要用于获取存储在MinIO中的对象的元数据信息，如对象的大小、最后修改时间等。
// 它通过调用MinIO的StatObject方法来实现。
//
// 参数:
//   - ctx: 上下文，用于传递请求的上下文信息，如超时设置。
//   - objectName: 对象名称，表示需要获取状态信息的对象在MinIO中的唯一标识。
//
// 返回值:
//   - minio.ObjectInfo: 包含对象状态信息的结构体，如对象的大小、最后修改时间等。
//   - error: 如果在获取对象状态信息过程中发生错误，则返回该错误。
func (s *MinIOStore) StatusObject(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	return s.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
}

func (s *MinIOStore) AbortUpload(ctx context.Context, objectName, uploadId string) error {
	return s.AbortMultipartUpload(ctx, bucket, objectName, uploadId)
}
