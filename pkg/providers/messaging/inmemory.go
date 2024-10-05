package messaging

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	Headers    map[string]string
	RoutingKey string
	Body       interface{}
}

func (b *InMemoryMessageBus) Publish(ctx context.Context, message interface{}) {
	routingKey := reflect.TypeOf(message).String()
	propagator := otel.GetTextMapPropagator()
	headers := make(map[string]string)
	propagator.Inject(ctx, propagation.MapCarrier(headers))
	tracer := otel.GetTracerProvider().Tracer("inmemory")
	ctx, span := tracer.Start(ctx, fmt.Sprintf("publish %s", routingKey),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("messaging.system", "inmemory")),
		trace.WithAttributes(attribute.String("messaging.destination.name", routingKey)))

	defer span.End()
	b.logger.Info("Received message",
		zap.String("type", routingKey))

	b.messageQueue <- Message{RoutingKey: routingKey, Body: message, Headers: headers}
}

func (b *InMemoryMessageBus) RegisterConsumer(messageType reflect.Type, consumer ConsumerFunc) {
	sugar := b.logger.Sugar()
	routingKey := messageType.String()
	sugar.Infof("Registering consumer for %s", routingKey)
	b.consumers[routingKey] = consumer
}

func (b *InMemoryMessageBus) Start() {
	go func() {
		for msg := range b.messageQueue {
			func() {
				// TODO: should probably use some consts here
				tracer := otel.GetTracerProvider().Tracer("inmemory")
				propagator := otel.GetTextMapPropagator()
				ctx := context.Background()
				ctx = propagator.Extract(ctx, propagation.MapCarrier(msg.Headers))
				ctx, span := tracer.Start(ctx, fmt.Sprintf("receive %s", msg.RoutingKey),
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(attribute.String("messaging.system", "inmemory")),
					trace.WithAttributes(attribute.String("messaging.destination.name", msg.RoutingKey)))
				defer span.End()
				if consumer, ok := b.consumers[msg.RoutingKey]; ok {
					if err := consumer(ctx, msg.Body); err != nil {
						span.SetStatus(codes.Error, "consuming message failed")
						span.RecordError(err)
					}
				}
			}()
		}
	}()
}

func (b *InMemoryMessageBus) Stop() {
	b.logger.Info("Stopping in memory message bus")
	close(b.messageQueue)
}
