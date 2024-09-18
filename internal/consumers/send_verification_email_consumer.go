package consumers

import (
	"go.uber.org/zap"
	"reflect"
)

type SendVerificationEmailConsumer struct {
}

func NewSendVerificationEmailConsumer() *SendVerificationEmailConsumer {
	return &SendVerificationEmailConsumer{}
}

func (c *SendVerificationEmailConsumer) Handle(message interface{}) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	routingKey := reflect.TypeOf(message).String()
	logger.Info("Received message in consumer",
		zap.String("type", routingKey))
}
