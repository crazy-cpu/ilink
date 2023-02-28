package nng

import (
	"fmt"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	"log"
)

type Sub struct{ SubSocket mangos.Socket }

func NewSub(url string) *Sub {
	var subSock mangos.Socket
	var err error
	if subSock, err = sub.NewSocket(); err != nil {
		log.Println(err)
	}
	subSock.SetOption(mangos.OptionSubscribe, []byte(""))
	return &Sub{SubSocket: subSock}
}

func (S *Sub) Connect(url string) error {
	if S.SubSocket == nil {
		return fmt.Errorf("%s-%w", "nng Sub Connect()", ErrNullPointer)
	}

	if err := S.SubSocket.Dial(url); err != nil {
		return err
	}
	return nil

}

//func (S *Sub) Publish(data []byte) Token {
//	var token Token
//	if S.SubSocket == nil {
//		token.err = fmt.Errorf("%s-%w", "nng Publish()", ErrNullPointer)
//	}
//	err := S.SubSocket.Send(data)
//	if err != nil {
//		token.err = err
//	}
//	return token
//}

func (S *Sub) Subscribe(handler callback) Token {
	var token Token
	if S.SubSocket == nil {
		token.err = fmt.Errorf("%s-%w", "nng Subscribe()", ErrNullPointer)
	}
	go func() {
		for {
			B, err := S.SubSocket.Recv()
			if err != nil {
				log.Println(err)
			}

			m := Message{data: B}

			handler(m)
		}
	}()

	return token
}
