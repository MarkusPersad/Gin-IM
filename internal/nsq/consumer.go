package nsq

import (
	"Gin-IM/pkg/protocol"
	"os"
	"strconv"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

var consumer *nsq.Consumer

func init() {
	go func() {
		for {
			if consumer == nil {
				consumer = newConsumer()
			} else {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

// newConsumer 创建并初始化 NSQ 消费者
func newConsumer() *nsq.Consumer {
	config := nsq.NewConfig()
	if maxInFlight, err := strconv.Atoi(os.Getenv("NSQ_MAX_INFLIGHT")); err == nil {
		config.MaxInFlight = maxInFlight
	}
	if consumers, err := nsq.NewConsumer(os.Getenv("NSQ_TOPIC"), os.Getenv("NSQ_CHANNEL"), config); err != nil {
		log.Logger.Error().Err(err).Msg("failed to create nsq consumer")
		return nil
	} else {
		consumers.AddHandler(nsq.HandlerFunc(handleMessage))
		if err := consumer.ConnectToNSQLookupd(os.Getenv("NSQ_ADDRESS")); err != nil {
			log.Logger.Fatal().Err(err).Msg("failed to connect to nsqlookupd")
			return nil
		}
		return consumers
	}
}

// handleMessage 处理从 NSQ 主题接收到的消息
func handleMessage(m *nsq.Message) error {
	var message protocol.Message
	if err := proto.Unmarshal(m.Body, &message); err != nil {
		log.Logger.Error().Err(err).Msg("failed to unmarshal message")
		return err
	}
	//TODO发送消息
	log.Logger.Info().Msgf("received message: %v", message.String())
	return nil
}

// StopConsumer 停止 NSQ 消费者
func StopConsumer() {
	if consumer != nil {
		consumer.Stop()
		log.Logger.Info().Msg("nsq consumer stopped")
	}
}
