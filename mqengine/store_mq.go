package mqengine

import (
	"github.com/streadway/amqp"
)

func NewStoreMQ(productId string) *RabbitMq {
	return &RabbitMq{
		queueName:    productId + config.RabbitMq.StoreQueue.QueueSuffix,
		routeKey:     productId + config.RabbitMq.StoreQueue.RouteKeySuffix,
		exchangeName: productId + config.RabbitMq.StoreQueue.ExchangeSuffix,
		exchangeType: amqp.ExchangeDirect,
	}
}
