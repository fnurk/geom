package pubsub

import ()

type Pubsub interface {
	Subscribe(pattern string, handler func(*Message, Subscriber), shutdownHandler func()) Subscriber
	Publish(msg *Message)
	Shutdown()
}

type Subscriber interface {
	Unsubscribe()
}

type Message struct {
	Topic string
	Body  string
}
