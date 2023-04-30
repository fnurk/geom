package pubsub

import (
	"github.com/tidwall/match"
	"sync"
)

type RBSubscriber struct {
	TopicPattern string
	Active       bool
	Pubsub       *RBPubsub
	Cursor       uint64
	Mutex        *sync.RWMutex
}

type RBPubsub struct {
	Active bool
	Buffer []*Message
	Cursor uint64
	Length uint64
	lock   sync.RWMutex
	Cond   *sync.Cond
}

func NewRingBufferPubsub(len uint64) Pubsub {
	p := RBPubsub{
		Cursor: 0,
		Length: len,
		Buffer: make([]*Message, len),
		lock:   sync.RWMutex{},
		Active: true,
	}
	p.Cond = &sync.Cond{}
	p.Cond.L = &p.lock
	return &p
}

func step(c uint64, l uint64) uint64 {
	return (c + 1) % l
}

func (p *RBPubsub) Subscribe(pattern string, handler func(*Message, Subscriber), shutdownHandler func()) Subscriber {
	s := RBSubscriber{
		TopicPattern: pattern,
		Active:       true,
		Pubsub:       p,
		Cursor:       p.Cursor,
	}

	go func(s *RBSubscriber) {
		p.lock.Lock()
		defer p.lock.Unlock()
		for {
			for p.Cursor == s.Cursor {
				if !s.Active || !p.Active {
					s.Active = false
					shutdownHandler()
					return
				}
				p.Cond.Wait()
			}
			if !s.Active || !p.Active {
				s.Active = false
				shutdownHandler()
				return
			}
			s.Cursor = step(s.Cursor, p.Length)
			if match.Match(p.Buffer[s.Cursor].Topic, s.TopicPattern) {
				handler(p.Buffer[s.Cursor], s)
			}
		}
	}(&s)

	return s
}

func (p *RBPubsub) Publish(msg *Message) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Cursor = step(p.Cursor, p.Length)
	p.Buffer[p.Cursor] = msg
	p.Cond.Broadcast()
}

func (p *RBPubsub) Shutdown() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Active = false
	p.Cond.Broadcast()
}

func (s RBSubscriber) Unsubscribe() {
	s.Active = false
}
