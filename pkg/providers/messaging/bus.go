package messaging

import "reflect"

type MessageBus interface {
	Start()
	Stop()
	RegisterConsumer(reflect.Type, ConsumerFunc)
	Publish(interface{})
	//PublishBatch([]interface{})
}

type ConsumerFunc func(interface{}) error

type Consumer interface {
	Handle(interface{})
}
