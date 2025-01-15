package database

import (
	"context"
	"github.com/rs/zerolog/log"
)

type ValkeyService interface {
	SetAndTime(ctx context.Context, key, value string, timeout int64) error
	GetValue(ctx context.Context, key string) string
	DelValue(ctx context.Context, keys ...string) error
	SetListAndTime(ctx context.Context, key string, list []string, timeout int64) error
	AddToList(ctx context.Context, key string, element string) error
	GetList(ctx context.Context, key string) []string
}

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
func (s *service) SetAndTime(ctx context.Context, key, value string, timeout int64) error {
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
func (s *service) GetValue(ctx context.Context, key string) string {
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
func (s *service) DelValue(ctx context.Context, keys ...string) error {
	// 执行Del命令以删除指定的键
	if err := s.valClient.Do(ctx, s.valClient.B().Del().Key(keys...).Build()).Error(); err != nil {
		// 如果发生错误，记录错误日志并返回错误
		log.Logger.Error().Err(err).Msg("valkey del error")
		// 返回错误
		return err
	}
	// 如果操作成功，返回nil
	return nil
}

// SetListAndTime 同时设置一个键对应的字符串列表值，并设置其过期时间。
// 该函数首先将一个字符串列表添加到指定的键中，然后为该键设置过期时间。
// 参数:
//
//	ctx: 上下文，用于传递请求范围的上下文信息。
//	key: 要设置的键。
//	list: 要添加到键的字符串列表。
//	timeout: 设置键的过期时间，单位为秒。
//
// 返回值:
//
//	如果成功设置列表和过期时间，则返回nil；否则返回错误。
func (s *service) SetListAndTime(ctx context.Context, key string, list []string, timeout int64) error {
	// 将字符串列表添加到键中
	if err := s.valClient.Do(ctx, s.valClient.B().Rpush().Key(key).Element(list...).Build()).Error(); err == nil {
		// 设置键的过期时间
		return s.valClient.Do(ctx, s.valClient.B().Expire().Key(key).Seconds(timeout).Build()).Error()
	} else {
		// 返回添加字符串列表时发生的错误
		return err
	}
}

// AddToList 将元素添加到列表中。
// 如果列表不存在，将创建一个列表并添加元素。
// 参数:
//
//	ctx - 上下文，用于传递请求范围的 deadline、取消信号等。
//	key - 列表的键。
//	element - 要添加到列表的元素。
//
// 返回值:
//
//	如果操作成功，返回 nil；否则返回错误。
func (s *service) AddToList(ctx context.Context, key string, element string) error {
	// 使用 valClient 执行 Rpushx 操作，将元素添加到列表的末尾。
	// Rpushx 命令仅在列表已存在时才执行添加操作。
	// 如果列表不存在，命令将失败并不进行任何操作。
	return s.valClient.Do(ctx, s.valClient.B().Rpushx().Key(key).Element(element).Build()).Error()
}

// GetList 获取给定键对应的列表值。
// 该方法从valClient中获取与键key相关联的列表值，并将其作为字符串切片返回。
// 参数:
//
//	ctx - 上下文，用于取消请求和传递请求级值。
//	key - 要获取列表值的键。
//
// 返回值:
//
//	字符串切片，包含键key对应的列表值。
//	如果键不存在或获取值时发生错误，则返回nil。
func (s service) GetList(ctx context.Context, key string) []string {
	// 执行缓存获取操作，使用Lrange命令获取列表的全部元素。
	result := s.valClient.Do(ctx, s.valClient.B().Lrange().Key(key).Start(0).Stop(-1).Build())

	// 检查操作是否成功，如果出现错误则记录错误日志并返回nil。
	if result.Error() != nil {
		log.Logger.Error().Err(result.Error()).Msg("valkey get list error")
		return nil
	}

	// 尝试将结果转换为字符串切片。
	if val, err := result.AsStrSlice(); err != nil {
		// 如果转换过程中出现错误，记录错误日志并返回nil。
		log.Logger.Error().Err(err).Msg("valkey get list error")
		return nil
	} else {
		// 转换成功，返回字符串切片。
		return val
	}
}
