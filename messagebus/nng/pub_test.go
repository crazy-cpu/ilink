package nng

import (
	"testing"
	"time"
)

func date() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func TestNNG(t *testing.T) {
	Pub := NewPub("tcp://127.0.0.1:40889")
	Sub := NewSub("tcp://127.0.0.1:40889")
	if err := Sub.Connect("tcp://127.0.0.1:40889"); err != nil {
		t.Error(err)
		return
	}

	if token := Sub.Subscribe(func(message Message) {
		t.Logf("reci1:%s", message)
	}); token.err != nil {
		t.Error(token.err)
	}
	if token := Pub.Publish([]byte(date())); token.err != nil {
		t.Error(token.err)
	}
	time.Sleep(time.Second)
}
