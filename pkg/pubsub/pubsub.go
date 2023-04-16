package pubsub

import (
	"sync"
)

type Message struct {
	Body string
}

type Subscriber struct {
	Topic       string
	MessageChan chan Message
	CloseChan   chan struct{}
	Active      bool
	Pubsub      Pubsub
}

type Pubsub struct {
	Subscriptions map[string][]Subscriber
	Mutex         *sync.RWMutex
}

func New() Pubsub {
	return Pubsub{
		Mutex:         &sync.RWMutex{},
		Subscriptions: make(map[string][]Subscriber),
	}
}

func (p Pubsub) Subscribe(topic string) Subscriber {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	sub := Subscriber{
		Topic:       topic,
		MessageChan: make(chan Message),
		Active:      true,
		Pubsub:      p,
	}
	p.Subscriptions[topic] = append(p.Subscriptions[topic], sub)
	return sub
}

func (p Pubsub) Publish(topic string, msg Message) {
	p.Mutex.RLock()
	defer p.Mutex.RUnlock()
	subscribers := p.Subscriptions[topic]
	for _, sub := range subscribers {
		if sub.Active {
			sub.MessageChan <- msg
		}
	}
}

func (s Subscriber) Unsubscribe() {
	s.Active = false
	s.Pubsub.Mutex.Lock()
	defer s.Pubsub.Mutex.Unlock()

	n := 0
	for _, sub := range s.Pubsub.Subscriptions[s.Topic] {
		if sub.Active {
			s.Pubsub.Subscriptions[s.Topic][n] = sub
			n++
		}
	}

	s.Pubsub.Subscriptions[s.Topic] = s.Pubsub.Subscriptions[s.Topic][:n]

}
