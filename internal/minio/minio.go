package minio

import (
	"Gin-IM/pkg/defines"
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
	"time"
)

func init() {
	if bucket == "" {
		if buckets := os.Getenv("MINIO_BUCKET_NAME"); buckets != "" {
			bucket = buckets
		} else {
			bucket = defines.DEFAUT_BUCKETNAME
		}
	}
}

type MinIOStore struct {
	*minio.Client
	*minio.Core
}

var (
	minIOStore *MinIOStore
	once       sync.Once
	bucket     string
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
				Core:   &minio.Core{Client: client},
			}
			log.Info().Msg("minio client init success")
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

	// 桶成功创建并设置策略，返回 nil 表示操作成功。
	return nil
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
func (s *MinIOStore) GetFileSign(ctx context.Context, objectName string) (string, error) {
	// 使用s.PresignedGetObject方法生成文件的预签名URL
	// 它允许在定义的时间内（此处为1小时）访问指定的文件对象
	if preSignedURL, err := s.PresignedGetObject(ctx, bucket, objectName, defines.FILE_SHORT_SIGN*time.Hour, nil); err != nil {
		// 如果生成预签名URL时发生错误，则记录错误日志并返回错误
		log.Logger.Error().Err(err).Msgf("failed to get file sign,%s/%s", bucket, objectName)
		return "", err
	} else {
		// 如果成功生成预签名URL，则返回URL字符串和nil错误
		return preSignedURL.String(), nil
	}
}

// DeleteFiles 删除指定的文件。
//
// ctx: 上下文，用于传递请求、超时等信息。
// forceDelete: 是否强制删除被治理策略保护的对象。
// objectNames: 要删除的对象名称列表。
//
// 返回错误，如果删除过程中遇到任何问题。
func (s *MinIOStore) DeleteFiles(ctx context.Context, forceDelete bool, objectNames ...string) error {
	// 创建一个通道，用于传递要删除的对象信息。
	objectsCh := make(chan minio.ObjectInfo)

	// 启动一个goroutine，将所有要删除的对象名称发送到通道中。
	go func() {
		defer close(objectsCh)
		for _, value := range objectNames {
			objectsCh <- minio.ObjectInfo{Key: value}
		}
	}()

	// 用于收集删除过程中遇到的错误。
	var errs []error

	// 调用RemoveObjects方法开始删除对象，它会返回一个错误通道。
	errCh := s.RemoveObjects(ctx, bucket, objectsCh, minio.RemoveObjectsOptions{GovernanceBypass: forceDelete})

	// 遍历错误通道，记录并收集删除失败的对象错误。
	for errDel := range errCh {
		log.Logger.Error().Err(errDel.Err).Msgf("failed to Delete %s", errDel.ObjectName)
		errs = append(errs, errDel.Err)
	}

	// 如果有对象删除失败，则返回一个汇总错误。
	if len(errs) > 0 {
		return errors.New("some objects failed to delete")
	}

	// 所有对象删除成功，返回nil。
	return nil
}
