package pubsub

import (
	"sync"
)

type Message struct {
	Body string
}

type Subscriber struct {
	Topic  string
	Active bool
	Pubsub *Pubsub
	Cursor uint64
	Mutex  *sync.RWMutex
}

type Pubsub struct {
	Active bool
	Buffer []Message
	Cursor uint64
	Length uint64
	lock   sync.RWMutex
	Cond   *sync.Cond
}

func New(len uint64) *Pubsub {
	p := &Pubsub{
		Cursor: 0,
		Length: len,
		Buffer: make([]Message, len),
		lock:   sync.RWMutex{},
		Active: true,
	}
	p.Cond = &sync.Cond{}
	p.Cond.L = &p.lock
	return p
}

func step(c uint64, l uint64) uint64 {
	return (c + 1) % l
}

func (p *Pubsub) Subscribe(topic string, handler func(*Subscriber, Message), shutdownHandler func()) *Subscriber {
	s := Subscriber{
		Topic:  topic,
		Active: true,
		Pubsub: p,
		Cursor: p.Cursor,
	}

	go func(s *Subscriber) {
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
			handler(s, p.Buffer[s.Cursor])
		}
	}(&s)

	return &s
}

func (p *Pubsub) Publish(topic string, msg Message) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Cursor = step(p.Cursor, p.Length)
	p.Buffer[p.Cursor] = msg
	p.Cond.Broadcast()
}

func (p *Pubsub) Shutdown() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Active = false
	p.Cond.Broadcast()
}

func (s *Subscriber) Unsubscribe() {
	s.Active = false
}
