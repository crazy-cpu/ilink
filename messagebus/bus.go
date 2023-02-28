package message

type Bus interface {
	Connect()

	Publish(arg interface{}) error

	Subscribe(topic string) error
}
