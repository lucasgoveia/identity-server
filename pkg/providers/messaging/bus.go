package messaging

type MessageBus interface {
	Publish(interface{})
	//PublishBatch([]interface{})
}
