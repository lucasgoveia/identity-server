package messaging

import (
	"go.uber.org/zap"
	"reflect"
)

type InMemoryMessageBus struct {
}

func (*InMemoryMessageBus) Publish(message interface{}) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	logger.Info("Received message",
		zap.String("type", reflect.TypeOf(message).String()))
}
