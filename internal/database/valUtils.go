package database

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SetAndTime 是一个方法，用于在服务中设置一个具有过期时间的键值对。
// 这个方法通过 valClient 执行 Setex 命令，将给定的 key 与 value 关联，并设置过期时间（秒）。
// 参数:
//
//	ctx: *gin.Context 类型，通常用于处理 HTTP 请求的上下文。
//	key: string 类型，表示要设置的键。
//	value: string 类型，表示与键关联的值。
//	timeout: int64 类型，表示键的过期时间，单位为秒。
//
// 返回值:
//
//	error 类型，如果执行操作过程中发生错误，则返回该错误。
func (s *service) SetAndTime(ctx *gin.Context, key, value string, timeout int64) error {
	// 使用 valClient 执行 Setex 命令，设置键值对和过期时间。
	return s.valClient.Do(ctx, s.valClient.B().Setex().Key(key).Seconds(timeout).Value(value).Build()).Error()
}

// GetValue 通过键值获取对应的值。
// 该方法使用 valClient 执行获取值的操作，主要执行以下步骤：
// 1. 使用传入的上下文和键值构建并发送一个获取值的请求。
// 2. 检查请求执行结果是否有错误，如果有错误则返回空字符串。
// 3. 将结果转换为字符串并返回，如果转换过程中出现错误，则记录错误日志并返回空字符串。
// 参数:
//
//	ctx *gin.Context - HTTP 请求的上下文，包含请求、响应和路由信息。
//	key string - 需要获取值的键。
//
// 返回值:
//
//	string - 键对应的值，如果获取失败或转换失败则返回空字符串。
func (s *service) GetValue(ctx *gin.Context, key string) string {
	// 执行获取值的操作，并检查是否有错误发生。
	result := s.valClient.Do(ctx, s.valClient.B().Get().Key(key).Build())
	if result.Error() != nil {
		return ""
	}

	// 将获取到的结果转换为字符串，如果转换过程中出现错误，则记录错误日志并返回空字符串。
	val, err := result.ToString()
	if err != nil {
		log.Logger.Error().Err(err).Msg("valkey get error")
		return ""
	}

	// 返回转换后的值。
	return val
}

// DelValue 删除指定的键值对
// 该方法使用Redis命令Del来删除给定的key
// 参数:
//
//	ctx *gin.Context - Gin框架的上下文，用于处理HTTP请求和响应
//	key string - 要删除的键的名称
//
// 返回值:
//
//	error - 如果删除操作失败，则返回错误；否则返回nil
func (s *service) DelValue(ctx *gin.Context, key string) error {
	// 执行Del命令以删除指定的键
	if err := s.valClient.Do(ctx, s.valClient.B().Del().Key(key).Build()).Error(); err != nil {
		// 如果发生错误，记录错误日志并返回错误
		log.Logger.Error().Err(err).Msg("valkey del error")
		return err
	}
	// 如果操作成功，返回nil
	return nil
}
