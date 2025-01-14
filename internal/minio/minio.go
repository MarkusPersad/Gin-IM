package minio

import (
	"Gin-IM/pkg/defines"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

func init() {
	once.Do(func() {
		if maxFileSize == 0 {
			if maxSize, err := strconv.Atoi(os.Getenv("MINIO_MAX_SIZE")); err != nil {
				log.Logger.Error().Msg("minio max size error")
			} else {
				maxFileSize = int64(maxSize)
			}
		}
	})
}

type MinIOStore struct {
	*minio.Client
}

var (
	minIOStore  *MinIOStore
	once        sync.Once
	maxFileSize int64
)

// NewClient 创建并初始化一个新的 MinIO 客户端实例。
// 参数 useSSL 指定是否使用 SSL 加密连接。
// 返回值是初始化后的 MinIOStore 实例，如果初始化失败，则返回 nil。
func NewClient(useSSL bool) *MinIOStore {
	// 使用 once 确保 minIOStore 在整个程序生命周期内只被初始化一次。
	once.Do(func() {
		// 如果 minIOStore 已经被初始化，则直接返回，避免重复初始化。
		if minIOStore != nil {
			return
		}
		// 从环境变量中获取 MinIO 的端点、访问密钥 ID 和秘密访问密钥。
		endpoint := os.Getenv("MINIO_ENDPOINT")
		accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
		secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
		// 检查获取到的配置项是否完整，如果任一配置项缺失，则记录错误日志并返回。
		if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
			log.Error().Msg("minio config error")
			return
		}
		// 尝试使用获取到的配置项创建一个新的 MinIO 客户端实例。
		if client, err := minio.New(endpoint, &minio.Options{
			Secure: useSSL,
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		}); err != nil {
			// 如果创建过程中出现错误，则记录错误日志。
			log.Error().Err(err).Msg("minio client error")
		} else {
			// 如果创建成功，则初始化 minIOStore，并记录成功日志。
			minIOStore = &MinIOStore{
				Client: client,
			}
			log.Info().Msgf("minio client init success ,endpoint:%s", minIOStore.EndpointURL().String())
		}
	})
	// 返回初始化后的 minIOStore 实例。
	return minIOStore
}

// CreateBucket 创建一个桶，并指定其所在位置。如果桶已经存在或创建成功，则不会重复创建。
// 此外，该函数还会为新创建的桶设置公共读取权限策略。
//
// 参数:
//
//	ctx: 上下文，用于传递请求、超时等信息。
//	bucketName: 欲创建的桶的名称。
//	bucketLocation: 桶的地理位置。
//
// 返回值:
//
//	如果桶存在或成功创建并设置了策略，则返回 nil。
//	如果发生错误（如检查桶存在性失败、创建桶失败或设置策略失败），则返回错误信息。
func (s *MinIOStore) CreateBucket(ctx context.Context, bucketName, bucketLocation string) error {
	// 检查桶是否已经存在，以避免重复创建。
	exists, err := s.BucketExists(ctx, bucketName)
	if err != nil {
		// 如果检查桶存在性时发生错误，则记录错误并返回。
		log.Logger.Error().Err(err).Msg("minio bucket exists error")
		return err
	}
	// 如果桶已经存在，则无需创建，直接返回。
	if exists {
		return nil
	}
	// 尝试创建新桶，并指定其所在区域。
	if err := s.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: bucketLocation}); err != nil {
		// 如果创建桶时发生错误，则记录错误并返回。
		log.Logger.Error().Err(err).Msg("minio create bucket error")
		return err
	}
	// 定义公共读取权限策略，允许任何人获取桶中的对象。
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "PublicRead",
				"Effect": "Allow",
				"Principal": "*",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::` + bucketName + `/*"
			}
		]
	}`

	// 为新创建的桶设置公共读取权限策略。
	if err = s.SetBucketPolicy(ctx, bucketName, policy); err != nil {
		// 如果设置策略时发生错误，尝试删除刚创建的桶以保持一致性。
		log.Logger.Error().Err(err).Msg("minio set bucket policy error")
		if err := s.RemoveBucket(ctx, bucketName); err != nil {
			log.Logger.Error().Err(err).Msg("minio remove bucket error")
			return err
		}
		return err
	}

	// 桶成功创建并设置策略，返回 nil 表示操作成功。
	return nil
}

// UploadFile 上传文件到 MinIO 存储系统。
// 参数:
//
//	ctx: 上下文，用于传递请求的上下文信息。
//	file: 要上传的文件，作为 io.Reader 类型。
//	bucketName: 存储桶名称。
//	objectName: 对象名称，即文件在存储桶中的路径和名称。
//	fileSize: 文件大小，用于验证文件的合法性。
//
// 返回值:
//
//	返回上传文件的 URL 和可能出现的错误。
func (s *MinIOStore) UploadFile(ctx context.Context, file io.Reader, bucketName, objectName string, fileSize int64) (string, error) {
	// 验证 fileSize 的合法性
	if fileSize < 0 || fileSize > maxFileSize*1024*1024*1024 {
		log.Logger.Error().Str("bucketName", bucketName).Str("objectName", objectName).Int64("fileSize", fileSize).Msg("invalid file size")
		return "", fmt.Errorf("invalid file size: %d", fileSize)
	}
	// 创建Bucket
	if err := s.CreateBucket(ctx, bucketName, ""); err != nil {
		return "", err
	}

	// 尝试将 file 转换为 io.Closer
	closer, ok := file.(io.Closer)
	if ok {
		defer func() {
			if err := closer.Close(); err != nil {
				log.Logger.Warn().Err(err).Msg("failed to close file")
			}
		}()
	}

	// 执行上传操作
	if _, err := s.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{}); err != nil {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Logger.Error().Err(err).Str("bucketName", bucketName).Str("objectName", objectName).Int64("fileSize", fileSize).Msg("context canceled or deadline exceeded")
			return "", ctx.Err()
		}
		log.Logger.Error().Err(err).Str("bucketName", bucketName).Str("objectName", objectName).Int64("fileSize", fileSize).Msg("failed to upload file to MinIO")
		return "", err
	}

	// 构建安全的 URL
	u, err := url.Parse(s.EndpointURL().String())
	if err != nil {
		log.Logger.Error().Err(err).Str("endpointURL", s.EndpointURL().String()).Msg("failed to parse endpoint URL")
		return "", err
	}
	u.Path = path.Join(u.Path, bucketName, objectName)

	return u.String(), nil
}

// GetFileSign 获取文件的预签名URL
// 该方法用于生成一个带有签名的文件URL，以便在未 来的特定时间内访问对象存储中的文件
// 参数:
//
//	ctx context.Context: 上下文信息，用于传递请求相关的配置或信息
//	bucketName string: 存储桶名称，标识文件所在的存储区域
//	objectName string: 文件对象名称，即文件在存储桶中的路径和名称
//
// 返回值:
//
//	string: 文件的预签名URL，允许在限定时间内访问文件
//	error: 错误信息，如果生成预签名URL时发生错误，则返回该错误
func (s *MinIOStore) GetFileSign(ctx context.Context, bucketName, objectName string) (string, error) {
	// 使用s.PresignedGetObject方法生成文件的预签名URL
	// 它允许在定义的时间内（此处为1小时）访问指定的文件对象
	if preSignedURL, err := s.PresignedGetObject(ctx, bucketName, objectName, defines.FILE_SHORT_SIGN*time.Hour, nil); err != nil {
		// 如果生成预签名URL时发生错误，则记录错误日志并返回错误
		log.Logger.Error().Err(err).Msgf("failed to get file sign,%s/%s", bucketName, objectName)
		return "", err
	} else {
		// 如果成功生成预签名URL，则返回URL字符串和nil错误
		return preSignedURL.String(), nil
	}
}
