package fakes

import (
	"context"
	"identity-server/pkg/providers/messaging"
	"reflect"
)

type FakeMessageBus struct {
	messages  []interface{}
	consumers map[string]messaging.ConsumerFunc
}

func (f *FakeMessageBus) Start() {}
func (f *FakeMessageBus) Stop()  {}
func (f *FakeMessageBus) RegisterConsumer(messageType reflect.Type, fn messaging.ConsumerFunc) {
	routingKey := messageType.String()
	f.consumers[routingKey] = fn
}
func (f *FakeMessageBus) Publish(ctx context.Context, message interface{}) {
	routingKey := reflect.TypeOf(message).String()
	headers := make(map[string]string)
	msg := messaging.Message{RoutingKey: routingKey, Body: message, Headers: headers}

	if consumer, ok := f.consumers[msg.RoutingKey]; ok {
		_ = consumer(ctx, msg.Body)
	}
}
