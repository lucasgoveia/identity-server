package accounts

import (
	"context"
	"go.uber.org/zap"
	"identity-server/pkg/providers"
	"identity-server/pkg/providers/messaging"
	"identity-server/tests/setup/dependencies"
	"os"
	"reflect"
	"testing"
)

var Deps *providers.DependencyContainer

type FakeMessageBus struct {
}

func (f *FakeMessageBus) Start()                                                                {}
func (f *FakeMessageBus) Stop()                                                                 {}
func (f *FakeMessageBus) RegisterConsumer(consumerType reflect.Type, fn messaging.ConsumerFunc) {}
func (f *FakeMessageBus) Publish(ctx context.Context, message interface{})                      {}

//Start()
//Stop()
//RegisterConsumer(reflect.Type, ConsumerFunc)
//Publish(ctx context.Context, message interface{})

func TestMain(m *testing.M) {
	appConfig, teardown := dependencies.SetupPostgresRedisConfig()

	Deps = providers.CreateDependencyContainer(appConfig)

	Deps.Bus = &FakeMessageBus{}
	Deps.Logger = zap.NewNop()

	code := m.Run()

	Deps.Destroy()
	teardown()
	os.Exit(code)
}
