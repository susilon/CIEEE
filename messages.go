package main

type Messages struct {	
	// sender client id
	Sid string
	// destination clientid
	Did string
	// messages
	Msg string
	// command type
	Cmd string
}