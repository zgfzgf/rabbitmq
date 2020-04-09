package mqengine

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"sync"
	"time"
)

type RabbitMq struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	queue        string
	routeKey     string
	exchange     string
	exchangeType string
	producerList []Producer
	receiverList []Receiver
	mu           sync.RWMutex
}

type QueueExchange struct {
	QueueName    string // 队列名
	RouteKey     string // 路由key
	ExchangeName string // 交换机名称
	ExchangeType string // 交换机类型
}

func (r *RabbitMq) connect() {
	url := fmt.Sprintf("amqp://%s:%s@%s/", config.RabbitMq.User, config.RabbitMq.Password, config.RabbitMq.Addr)
	logger.Debug(url)
	if conn, err := amqp.Dial(url); err != nil {
		logger.Error("消息队列连接失败!",
			zap.String("url", url),
			zap.Error(err))
		return
	} else {
		logger.Info("Success connect rabbitmq",
			zap.String("url", url))
		r.connection = conn
	}

	if channel, err := r.connection.Channel(); err != nil {
		logger.Error("消息队列打开channel失败!",
			zap.Error(err))
		return
	} else {
		logger.Info("Success open channel",
			zap.String("url", url))
		r.mu.Lock()
		r.channel = channel
		r.mu.Unlock()
	}
}

func (r *RabbitMq) close() {
	if err := r.channel.Close(); err != nil {
		logger.Warn("close channel failed!", zap.Error(err))
		return
	}
	if err := r.connection.Close(); err != nil {
		logger.Warn("close amqp connection failed!", zap.Error(err))
		return
	}
}

func NewAmqp(e *QueueExchange) *RabbitMq {
	return &RabbitMq{
		queue:        e.QueueName,
		routeKey:     e.RouteKey,
		exchange:     e.ExchangeName,
		exchangeType: e.ExchangeType,
	}
}
func (r *RabbitMq) Start() {
	r.connect()
	//defer r.close()
	for _, receiver := range r.receiverList {
		go r.listenReceiver(receiver)
	}

	for _, producer := range r.producerList {
		go r.listenProducer(producer)
	}
	time.Sleep(1 * time.Second)
}

func (r *RabbitMq) RegisterProducer(producer Producer) {
	r.producerList = append(r.producerList, producer)
}

func (r *RabbitMq) RegisterReceiver(receiver Receiver) {
	r.mu.Lock()
	r.receiverList = append(r.receiverList, receiver)
	r.mu.Unlock()
}
func (r *RabbitMq) listenProducer(producer Producer) {
	storeChan, ackChan := producer.Store()
	if r.channel == nil {
		r.connect()
	}

	if err := r.channel.ExchangeDeclare(r.exchange, r.exchangeType, true, false, false, false, nil); err != nil {
		logger.Error("declare exchange failed!",
			zap.String("exchange", r.exchange),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare exchange ",
			zap.String("exchange", r.exchange))
	}

	if _, err := r.channel.QueueDeclare(r.queue, true, false, false, false, nil); err != nil {
		logger.Error("declare channel failed!",
			zap.String("queue", r.queue),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare channel",
			zap.String("queue", r.queue))
	}

	if err := r.channel.QueueBind(r.queue, r.routeKey, r.exchange, false, nil); err != nil {
		logger.Error("bind channel failed!",
			zap.String("queue", r.queue),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchange),
			zap.Error(err))
		return
	} else {
		logger.Debug("Success bind channel",
			zap.String("queue", r.queue),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchange))
	}
	commit := true
	for {
		message, ok := <-storeChan
		if !ok && ackChan != nil {
			close(ackChan)
			logger.Info("ack channel close!")
			return
		}
		if err := r.channel.Publish(r.exchange, r.routeKey, false, false, amqp.Publishing{
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
		//for message := range storeChan {
		//	if err := r.channel.Publish(r.exchange, r.routeKey, false, false, amqp.Publishing{
		//		DeliveryMode:  amqp.Persistent,
		//		ContentType:   "text/plain",
		//		Body:          message.Body,
		//		Priority:      message.Priority,
		//		CorrelationId: message.CorrelationId,
		//		ReplyTo:       message.ReplyTo,
		//		MessageId:     message.MessageId,
		//		Timestamp:     message.Timestamp,
		//		Type:          message.Type,
		//		UserId:        message.UserId,
		//		AppId:         message.AppId,
		//	}); err != nil {
		//		logger.Error("mq publish message failed!", zap.Error(err))
		//		commit = false
		//		message.Commit = false
		//		ackChan <- message
		//	} else {
		//		logger.Info("Success publish message",
		//			zap.String("correlationId",message.CorrelationId),
		//			zap.ByteString("body",message.Body))
		//		switch message.Status {
		//		case MessageStatusStart:
		//			commit = true
		//		case MessageStatusEnd:
		//			if commit {
		//				message.Commit = true
		//				ackChan <- message
		//			}
		//		}
		//	}
	}
}

func (r *RabbitMq) listenReceiver(receiver Receiver) {
	readChan := receiver.Consumer()
	if r.channel == nil {
		r.connect()
	}

	if err := r.channel.ExchangeDeclare(r.exchange, r.exchangeType, true, false, false, false, nil); err != nil {
		logger.Error("declare exchange failed!",
			zap.String("exchange", r.exchange),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare exchange",
			zap.String("exchange", r.exchange))
	}

	if _, err := r.channel.QueueDeclare(r.queue, true, false, false, false, nil); err != nil {
		logger.Error("declare channel failed!",
			zap.String("queue", r.queue),
			zap.Error(err))
		return
	} else {
		logger.Info("Success declare channel",
			zap.String("queue", r.queue))
	}

	if err := r.channel.QueueBind(r.queue, r.routeKey, r.exchange, false, nil); err != nil {
		logger.Error("bind channel failed!",
			zap.String("queue", r.queue),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchange),
			zap.Error(err))
		return
	} else {
		logger.Debug("Success bind channel",
			zap.String("queue", r.queue),
			zap.String("routeKey", r.routeKey),
			zap.String("exchange", r.exchange))
	}

	r.channel.Qos(1, 0, true)
	if msgList, err := r.channel.Consume(r.queue, "", false, false, false, false, nil); err != nil {
		logger.Error("consume channel failed!",
			zap.String("queue", r.queue),
			zap.Error(err))
		return
	} else {
		for {
			msg := <-msgList
			message := NewMessage(&msg)
			readChan <- message
			logger.Debug("message",
				zap.String("messageId", message.MessageId),
				zap.String("type", message.Type),
				zap.Time("timestamp", message.Timestamp),
				zap.String("userId", message.UserId),
				zap.String("appId", message.AppId),
				zap.ByteString("body", message.Body))
		}
	}
}
