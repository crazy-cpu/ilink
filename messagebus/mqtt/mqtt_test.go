package mqtt

import (
	"fmt"
	emqx "github.com/eclipse/paho.mqtt.golang"
	"testing"
	"time"
)

func TestMqtt(t *testing.T) {
	M := &Client{
		Server:   "tcp://127.0.0.1:1883",
		ClientID: "testClient1",
	}
	M.Connect()
	if err := M.Publish(PubArg{PubQos: 0, PubRetained: false, PubTopic: "iot/device/test", PubPayload: []byte("hello,world!")}); err != nil {
		t.Error(err)
	}
	if err := M.Subscribe(SubArg{SubTopic: "iot/device/test", SubQos: 0, SubCallback: func(client emqx.Client, message emqx.Message) {
		fmt.Println(message)
	}}); err != nil {
		t.Error(err)
		return
	}
	time.Sleep(2 * time.Second)
}
