package mqengine

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"sync"
)

type RabbitMq struct {
	queueName    string // 队列名
	routeKey     string // 路由key
	exchangeName string // 交换机名称
	exchangeType string // 交换机类型
}

var connection *amqp.Connection
var connOnce sync.Once

func GetRabbitMqConn() *amqp.Connection {
	connOnce.Do(func() {
		url := fmt.Sprintf("amqp://%s:%s@%s/", config.RabbitMq.User, config.RabbitMq.Password, config.RabbitMq.Addr)
		logger.Debug(url)
		var err error
		connection, err = amqp.Dial(url)
		if err != nil {
			logger.Panic("rabbitMq connect failed!",
				zap.String("url", url),
				zap.Error(err))
		} else {
			logger.Info("Success rabbitMq connect ",
				zap.String("url", url))
		}
	})
	return connection
}
func CloseRabbitMqConn() {
	if err := connection.Close(); err != nil {
		logger.Warn("close connection failed!", zap.Error(err))
		return
	}
}

func NewChannel() *amqp.Channel {
	if channel, err := GetRabbitMqConn().Channel(); err != nil {
		logger.Panic("rabbitMq channel failed!",
			zap.Error(err))
		return nil
	} else {
		logger.Info("Success rabbitMq channel")
		return channel
	}
}
func CloseChannel(channel *amqp.Channel) {
	if err := channel.Close(); err != nil {
		logger.Warn("close channel failed!", zap.Error(err))
		return
	}
}

func (r *RabbitMq) Store(producer Producer) {
	storeChan, ackChan := producer.Store()
	channel := NewChannel()
	defer CloseChannel(channel)

	if err := channel.ExchangeDeclare(r.exchangeName, r.exchangeType, true, false, false, false, nil); err != nil {
		logger.Error("declare exchange failed!",
			zap.String("exchange", r.exchangeName),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare exchange ",
			zap.String("exchange", r.exchangeName))
	}

	if _, err := channel.QueueDeclare(r.queueName, true, false, false, false, nil); err != nil {
		logger.Error("declare channel failed!",
			zap.String("queue", r.queueName),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare channel",
			zap.String("queue", r.queueName))
	}

	if err := channel.QueueBind(r.queueName, r.routeKey, r.exchangeName, false, nil); err != nil {
		logger.Error("bind channel failed!",
			zap.String("queue", r.queueName),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchangeName),
			zap.Error(err))
		return
	} else {
		logger.Debug("Success bind channel",
			zap.String("queue", r.queueName),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchangeName))
	}
	commit := true
	for {
		message, ok := <-storeChan
		if !ok && ackChan != nil {
			close(ackChan)
			logger.Info("ack channel close!")
			return
		}
		if err := channel.Publish(r.exchangeName, r.routeKey, false, false, amqp.Publishing{
			DeliveryMode:  amqp.Persistent,
			ContentType:   "text/plain",
			Body:          message.Body,
			Priority:      message.Priority,
			CorrelationId: message.CorrelationId,
			ReplyTo:       message.ReplyTo,
			MessageId:     message.MessageId,
			Timestamp:     message.Timestamp,
			Type:          message.Type,
			UserId:        message.UserId,
			AppId:         message.AppId,
		}); err != nil {
			logger.Error("mq publish message failed!", zap.Error(err))
			commit = false
			message.Commit = false
			ackChan <- message
		} else {
			logger.Info("Success publish message",
				zap.String("correlationId", message.CorrelationId),
				zap.ByteString("body", message.Body))
			switch message.Status {
			case MessageStatusStart:
				commit = true
			case MessageStatusEnd:
				if commit {
					message.Commit = true
					ackChan <- message
				}
			}
		}
	}
}

func (r *RabbitMq) Reader(ctx context.Context, receiver Receiver) {
	readChan := receiver.Reader()
	channel := NewChannel()
	defer CloseChannel(channel)

	if err := channel.ExchangeDeclare(r.exchangeName, r.exchangeType, true, false, false, false, nil); err != nil {
		logger.Error("declare exchange failed!",
			zap.String("exchange", r.exchangeName),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare exchange",
			zap.String("exchange", r.exchangeName))
	}

	if _, err := channel.QueueDeclare(r.queueName, true, false, false, false, nil); err != nil {
		logger.Error("declare channel failed!",
			zap.String("queue", r.queueName),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare channel",
			zap.String("queue", r.queueName))
	}

	if err := channel.QueueBind(r.queueName, r.routeKey, r.exchangeName, false, nil); err != nil {
		logger.Error("bind channel failed!",
			zap.String("queue", r.queueName),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchangeName),
			zap.Error(err))
		return
	} else {
		logger.Debug("Success bind channel",
			zap.String("queue", r.queueName),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchangeName))
	}

	channel.Qos(1, 0, true)
	if msgList, err := channel.Consume(r.queueName, "", false, false, false, false, nil); err != nil {
		logger.Error("consume channel failed!",
			zap.String("queue", r.queueName),
			zap.Error(err))
		return
	} else {
		for {
			select {
			case msg := <-msgList:
				message := NewMessage(&msg)
				readChan <- message
				logger.Debug("message",
					zap.String("messageId", message.MessageId),
					zap.String("type", message.Type),
					zap.Time("timestamp", message.Timestamp),
					zap.String("userId", message.UserId),
					zap.String("appId", message.AppId),
					zap.ByteString("body", message.Body))
			case <-ctx.Done():
				close(readChan)
				logger.Info("close readChan")
				return
			}
		}
	}
}
