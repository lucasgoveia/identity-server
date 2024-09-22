package messaging

import (
	"go.uber.org/zap"
	"reflect"
)

type InMemoryMessageBus struct {
	consumers    map[string]ConsumerFunc
	messageQueue chan Message
	logger       *zap.Logger
}

func NewInMemoryMessageBus(logger *zap.Logger) *InMemoryMessageBus {
	return &InMemoryMessageBus{
		consumers:    make(map[string]ConsumerFunc),
		messageQueue: make(chan Message),
		logger:       logger,
	}
}

// Register a map of consumer names to functions

// Message struct with routing key and payload
type Message struct {
	RoutingKey string
	Body       interface{}
}

func (b *InMemoryMessageBus) Publish(message interface{}) {
	routingKey := reflect.TypeOf(message).String()
	b.logger.Info("Received message",
		zap.String("type", routingKey))

	b.messageQueue <- Message{RoutingKey: routingKey, Body: message}
}

func (b *InMemoryMessageBus) RegisterConsumer(messageType reflect.Type, consumer ConsumerFunc) {
	sugar := b.logger.Sugar()
	routingKey := messageType.String()
	sugar.Infof("Registering consumer for %s", routingKey)
	b.consumers[routingKey] = consumer
}

func (b *InMemoryMessageBus) routeMessage(msg Message) {
	sugar := b.logger.Sugar()
	consumerFunc, ok := b.consumers[msg.RoutingKey]
	if ok {
		consumerFunc(msg.Body)
	} else {
		sugar.Errorf("No consumer registered for %s", msg.RoutingKey)
	}
}

func (b *InMemoryMessageBus) Start() {
	go func() {
		for msg := range b.messageQueue {
			if consumer, ok := b.consumers[msg.RoutingKey]; ok {
				consumer(msg.Body)
			}
		}
	}()
}

func (b *InMemoryMessageBus) Stop() {
	b.logger.Info("Stopping in memory message bus")
	close(b.messageQueue)
}
