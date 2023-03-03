package mqtt

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"reflect"
)

var (
	illegalArg = fmt.Errorf("%s", "非法入参")
)

type Client struct {
	Handler  mqtt.Client
	Server   string
	ClientID string
	UserName string
	Password string
}

type PubArg struct {
	PubQos      byte
	PubTopic    string
	PubRetained bool
	PubPayload  []byte
}

type PubArgOption func(arg *PubArg)

type SubArg struct {
	SubTopic    string
	SubQos      byte
	SubCallback mqtt.MessageHandler
}

func (C *Client) Connect() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(C.Server)
	opts.SetClientID(C.ClientID)
	opts.SetUsername(C.UserName)
	opts.SetPassword(C.Password)

	C.Handler = mqtt.NewClient(opts)
	if token := C.Handler.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func (C *Client) Publish(args any) error {
	if reflect.TypeOf(args) != reflect.TypeOf(PubArg{}) {
		return errors.Wrap(illegalArg, "Mqtt Publish()")
	}
	arg := args.(PubArg)
	if Token := C.Handler.Publish(arg.PubTopic, arg.PubQos, arg.PubRetained, arg.PubPayload); Token.Wait() && Token.Error() != nil {
		return errors.Wrap(Token.Error(), "mqtt Publish()")
	}
	return nil
}

func (C *Client) Subscribe(args any) error {
	if reflect.TypeOf(args) != reflect.TypeOf(SubArg{}) {
		return errors.Wrap(illegalArg, "Mqtt Subscribe()")
	}
	arg := args.(SubArg)
	if token := C.Handler.Subscribe(arg.SubTopic, arg.SubQos, arg.SubCallback); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
