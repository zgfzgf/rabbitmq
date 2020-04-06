package mqengine

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

func NewStoreMQ(productId string /*, producer Producer*/) *RabbitMq {
	queueExchange := &QueueExchange{
		QueueName:    productId + config.RabbitMq.StoreQueue.QueueSuffix,
		RouteKey:     productId + config.RabbitMq.StoreQueue.RouteKeySuffix,
		ExchangeName: productId + config.RabbitMq.StoreQueue.ExchangeSuffix,
		ExchangeType: amqp.ExchangeDirect,
	}
	mq := NewAmqp(queueExchange)
	//mq.RegisterProducer(producer)
	logger.Info("Init StoreMQ",
		zap.String("queueName", queueExchange.QueueName),
		zap.String("routeKey", queueExchange.RouteKey),
		zap.String("exchangeName", queueExchange.ExchangeName),
		zap.String("exchangeType", queueExchange.ExchangeType))
	return mq
}
