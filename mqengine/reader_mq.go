package mqengine

import (
	"github.com/streadway/amqp"
)

func NewReaderMQ(productId string) *RabbitMq {
	return &RabbitMq{
		queueName:    productId + config.RabbitMq.ReadQueue.QueueSuffix,
		routeKey:     productId + config.RabbitMq.ReadQueue.RouteKeySuffix,
		exchangeName: productId + config.RabbitMq.ReadQueue.ExchangeSuffix,
		exchangeType: amqp.ExchangeDirect,
	}
}
