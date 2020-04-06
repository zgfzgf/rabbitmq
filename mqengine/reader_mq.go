package mqengine

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func NewReaderMQ1(productId string, receiver Receiver) *RabbitMq {
	queueExchange := &QueueExchange{
		QueueName:    productId + config.RabbitMq.ReadQueue.QueueSuffix,
		RouteKey:     productId + config.RabbitMq.ReadQueue.RouteKeySuffix,
		ExchangeName: productId + config.RabbitMq.ReadQueue.ExchangeSuffix,
		ExchangeType: amqp.ExchangeDirect,
	}
	mq := NewAmqp(queueExchange)
	mq.RegisterReceiver(receiver)
	logger.Info("Init ReaderMQ",
		zap.String("queueName", queueExchange.QueueName),
		zap.String("routeKey", queueExchange.RouteKey),
		zap.String("exchangeName", queueExchange.ExchangeName),
		zap.String("exchangeType", queueExchange.ExchangeType))
	return mq
}

func NewReaderMQ(productId string) *RabbitMq {
	queueExchange := &QueueExchange{
		QueueName:    productId + config.RabbitMq.ReadQueue.QueueSuffix,
		RouteKey:     productId + config.RabbitMq.ReadQueue.RouteKeySuffix,
		ExchangeName: productId + config.RabbitMq.ReadQueue.ExchangeSuffix,
		ExchangeType: amqp.ExchangeDirect,
	}
	mq := NewAmqp(queueExchange)
	logger.Info("Init ReaderMQ",
		zap.String("queueName", queueExchange.QueueName),
		zap.String("routeKey", queueExchange.RouteKey),
		zap.String("exchangeName", queueExchange.ExchangeName),
		zap.String("exchangeType", queueExchange.ExchangeType))
	return mq
}
