package nng

import (
	"fmt"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	// register transports
	_ "go.nanomsg.org/mangos/v3/transport/all"
	"log"
)

type Pub struct{ PubSocket mangos.Socket }

func NewPub(url string) *Pub {
	var pubSock mangos.Socket
	var err error
	//pub
	if pubSock, err = pub.NewSocket(); err != nil {
		log.Println(err)
	}
	if err = pubSock.Listen(url); err != nil {
		log.Println(err)
	}

	return &Pub{PubSocket: pubSock}

}

func (P *Pub) Connect() {}

func (P *Pub) Publish(data []byte) Token {
	var token Token
	if P.PubSocket == nil {
		token.err = fmt.Errorf("%s-%w", "nng Publish()", ErrNullPointer)
	}
	err := P.PubSocket.Send(data)
	fmt.Println("sendData", string(data))
	if err != nil {
		token.err = err
	}
	return token
}

//
//func (P *Pub) Subscribe(handler callback) Token {
//	var token Token
//	if P.PubSocket == nil {
//		token.err = fmt.Errorf("%s-%w", "nng Subscribe()", ErrNullPointer)
//	}
//	go func() {
//		for {
//			B, err := P.PubSocket.Recv()
//			if err != nil {
//				log.Println(err)
//			}
//
//			m := Message{data: B}
//
//			handler(m)
//		}
//	}()
//
//	return token
//}
