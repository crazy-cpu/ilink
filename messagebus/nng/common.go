package nng

import "errors"

var (
	ErrNullPointer = errors.New("nng Conn is null")
)

type Message struct{ data []byte }

type Token struct{ err error }

type callback func(Message)
