package mqengine

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func NewInfoMQ(productId string /*, producer Producer*/) *RabbitMq {
	queueExchange := &QueueExchange{
		QueueName:    productId + config.RabbitMq.InfoQueue.QueueSuffix,
		RouteKey:     productId + config.RabbitMq.InfoQueue.RouteKeySuffix,
		ExchangeName: productId + config.RabbitMq.InfoQueue.ExchangeSuffix,
		ExchangeType: amqp.ExchangeDirect,
	}
	mq := NewAmqp(queueExchange)
	// mq.RegisterProducer(producer)
	logger.Info("Init InfoMQ",
		zap.String("queueName", queueExchange.QueueName),
		zap.String("routeKey", queueExchange.RouteKey),
		zap.String("exchangeName", queueExchange.ExchangeName),
		zap.String("exchangeType", queueExchange.ExchangeType))
	return mq
}
