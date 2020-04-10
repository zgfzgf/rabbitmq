package mqengine

import (
	"github.com/streadway/amqp"
)

func NewInfoMQ(productId string) *RabbitMq {
	return &RabbitMq{
		queueName:    productId + config.RabbitMq.InfoQueue.QueueSuffix,
		routeKey:     productId + config.RabbitMq.InfoQueue.RouteKeySuffix,
		exchangeName: productId + config.RabbitMq.InfoQueue.ExchangeSuffix,
		exchangeType: amqp.ExchangeDirect,
	}
}
