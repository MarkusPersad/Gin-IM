package nsq

import (
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/protocol"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

var producer *nsq.Producer
var topic string

func init() {
	producer = newProducer()
	if producer == nil {
		log.Logger.Fatal().Msg("failed to create NSQ producer")
	}
	if err := producer.Ping(); err != nil {
		log.Logger.Fatal().Err(err).Msg("NSQ producer ping failed")
	}
	log.Logger.Info().Msg("NSQ producer initialized successfully")
}
func newProducer() *nsq.Producer {
	if producer != nil && producer.Ping() == nil {
		return producer
	}
	config := nsq.NewConfig()
	if maxInFlight, err := strconv.Atoi(os.Getenv("NSQ_MAX_INFLIGHT")); err == nil {
		config.MaxInFlight = maxInFlight
	}

	nsqAddr := os.Getenv("NSQ_ADDR")
	producer, err := nsq.NewProducer(nsqAddr, config)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("Failed to create producer")
	}

	topic = os.Getenv("NSQ_TOPIC")
	if topic == "" {
		topic = "chat"
	}
	return producer
}

func SendMsg(message *protocol.ChatMessage) error {
	// 创建一个通道以接收异步发布操作的结果。
	doneChan := make(chan *nsq.ProducerTransaction)

	// 将消息序列化为proto字节切片。
	if data, err := proto.Marshal(message); err != nil {
		// 如果序列化失败，记录错误并返回错误。
		log.Logger.Error().Err(err).Msg("failed to marshal message")
		return err
	} else {
		// 异步发布序列化后的消息数据到NSQ主题。
		if err := producer.PublishAsync(topic, data, doneChan); err != nil {
			// 如果发布失败，记录错误并返回错误。
			log.Logger.Error().Err(err).Msg("failed to publish message")
			return err
		}
	}

	// 等待发布操作完成或超时。
	select {
	case txn := <-doneChan:
		// 如果发布操作报告错误，记录错误并返回错误。
		if txn.Error != nil {
			log.Logger.Error().Err(txn.Error).Any("message", message).Msg("failed to publish message")
			return txn.Error
		}
	case <-time.After(defines.MESSAGE_SEND_TIMEOUT * time.Second):
		// 如果发布操作超时，记录错误并返回错误。
		log.Logger.Error().Any("message", message).Msg("Time out waiting to publish message")
		return errors.New("time out waiting to publish message")
	}
	return nil
}

func Close() {
	producer.Stop()
}
