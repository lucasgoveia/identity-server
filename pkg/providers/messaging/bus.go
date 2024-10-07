package messaging

import (
	"context"
	"reflect"
)

type MessageBus interface {
	Start()
	Stop()
	RegisterConsumer(reflect.Type, ConsumerFunc)
	Publish(ctx context.Context, message interface{})
}

type ConsumerFunc func(ctx context.Context, message interface{}) error

type Consumer interface {
	Handle(ctx context.Context, message interface{}) error
}
