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

// InitMutiparts 初始化分片上传的过程。
// 该方法根据上传信息生成上传ID和一系列预签名的上传URL，供客户端使用。
// 参数:
//   - ctx: 上下文，用于传递请求范围的数据和控制超时。
//   - upload: 包含上传信息的结构体，如分片大小、分片数量、对象名称等。
//
// 返回值:
//   - *types.UploadUrls: 包含上传ID和预签名URL的结构体指针。
//   - error: 在操作过程中可能遇到的错误。
func (s *MinIOStore) InitMutiparts(ctx context.Context, upload types.UploadInfo) (*types.UploadUrls, error) {
	if err := s.CreateBucket(ctx, bucket, ""); err != nil {
		return nil, err
	}
	// 初始化uploadUrls变量，用于存储上传ID和预签名URL。
	var uploadUrls types.UploadUrls

	// 检查分片大小和分片数量是否有效，如果无效则记录错误日志并返回错误。
	if upload.ChunkSize <= 0 || upload.ChunkNumber <= 0 {
		log.Logger.Error().Str("objectName", upload.ObjectName).Msg("invalid chunk size or chunk number")
		return nil, exception.ErrBadRequest
	}

	// 处理单个分片上传的情况。
	if upload.ChunkSize == 1 {
		// 生成单个分片上传的预签名URL。
		if presignedUrl, err := s.PresignedPutObject(ctx, bucket, upload.ObjectName, defines.FILE_SHORT_SIGN*time.Hour); err != nil {
			log.Logger.Error().Err(err).Msg("failed to get single upload url")
			return nil, err
		} else {
			uploadUrls.Urls = append(uploadUrls.Urls, presignedUrl.String())
			return &uploadUrls, nil
		}
	}

	// 检查上传ID是否为空，如果为空则发起新的分片上传请求。
	if upload.UploadId == "" {
		// 生成新的上传ID。
		if uploadId, err := s.NewMultipartUpload(ctx, bucket, upload.ObjectName, minio.PutObjectOptions{
			ContentType: upload.ContentType,
			PartSize:    upload.ChunkSize,
		}); err != nil {
			log.Logger.Error().Err(err).Str("objectName", upload.ObjectName).Msg("failed to create multipart upload")
			return nil, err
		} else {
			upload.UploadId = uploadId
		}

		// 设置uploadUrls的UploadId字段为新生成的上传ID。
		uploadUrls.UploadId = upload.UploadId

		// 初始化urlValues变量，用于存储查询参数。
		urlValues := url.Values{}
		urlValues.Set("uploadId", upload.UploadId)

		// 循环生成每个分片的预签名URL，并添加到uploadUrls的Urls列表中。
		for num := 1; num <= upload.ChunkNumber; num++ {
			urlValues.Set("partNumber", strconv.Itoa(num))
			if presignedUrl, err := s.Presign(ctx, "PUT", bucket, upload.ObjectName, defines.FILE_SHORT_SIGN*time.Hour, urlValues); err != nil {
				log.Logger.Error().Err(err).Msg("failed to get multipart upload url")
				return nil, err
			} else {
				uploadUrls.Urls = append(uploadUrls.Urls, presignedUrl.String())
			}
		}
	}

	// 返回包含上传ID和预签名URL的结构体指针。
	return &uploadUrls, nil
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
