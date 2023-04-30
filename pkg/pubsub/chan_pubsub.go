package pubsub

import (
	"sync"

	"github.com/tidwall/match"
)

type ChanPubsub struct {
	subMutex    *sync.RWMutex
	subscribers map[string]map[*ChanSubscriber]struct{}
	closeChan   chan struct{}
	active      bool
}

type ChanSubscriber struct {
	activeMutex *sync.RWMutex
	msgChan     chan *Message
	pubSub      *ChanPubsub
	active      bool
	pattern     string
}

func NewChanPubsub() Pubsub {
	p := ChanPubsub{
		subMutex:    &sync.RWMutex{},
		subscribers: map[string]map[*ChanSubscriber]struct{}{},
		closeChan:   make(chan struct{}),
		active:      true,
	}
	return &p
}

// Subscribe implements Pubsub
func (p *ChanPubsub) Subscribe(pattern string, handler func(*Message, Subscriber), shutdownHandler func()) Subscriber {
	p.subMutex.Lock()
	defer p.subMutex.Unlock()

	ch := make(chan *Message, 1)

	sub := ChanSubscriber{
		activeMutex: &sync.RWMutex{},
		msgChan:     ch,
		pubSub:      p,
		active:      true,
		pattern:     pattern,
	}

	if p.subscribers[pattern] == nil {
		p.subscribers[pattern] = make(map[*ChanSubscriber]struct{})
	}

	p.subscribers[pattern][&sub] = struct{}{}

	go func() {
		for msg := range sub.msgChan {
			handler(msg, &sub)
		}
	}()

	return &sub
}

// Publish implements Pubsub
func (p *ChanPubsub) Publish(msg *Message) {
	p.subMutex.RLock()
	defer p.subMutex.RUnlock()
	for k, v := range p.subscribers {
		if match.Match(msg.Topic, k) {
			for s := range v {
				s.msgChan <- msg
			}
		}
	}
}

// Shutdown implements Pubsub
func (p *ChanPubsub) Shutdown() {
	p.subMutex.Lock()
	defer p.subMutex.Unlock()
	for _, v := range p.subscribers {
		for s := range v {
			close(s.msgChan)
		}

	}
}

// Unsubscribe implements Subscriber
func (s *ChanSubscriber) Unsubscribe() {
	s.pubSub.subMutex.Lock()
	defer s.pubSub.subMutex.Unlock()
	delete(s.pubSub.subscribers[s.pattern], s)
	close(s.msgChan)
}
