package pubsub

import (
	"sync"

	"github.com/tidwall/match"
)

type Message struct {
	Topic string
	Body  string
}

type Subscriber struct {
	TopicPattern string
	Active       bool
	Pubsub       *Pubsub
	Cursor       uint64
	Mutex        *sync.RWMutex
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

func (p *Pubsub) Subscribe(pattern string, handler func(*Message, *Subscriber), shutdownHandler func()) *Subscriber {
	s := Subscriber{
		TopicPattern: pattern,
		Active:       true,
		Pubsub:       p,
		Cursor:       p.Cursor,
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
			if match.Match(p.Buffer[s.Cursor].Topic, s.TopicPattern) {
				handler(&p.Buffer[s.Cursor], s)
			}
		}
	}(&s)

	return &s
}

func (p *Pubsub) Publish(msg Message) {
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
	s.Pubsub.Cond.Broadcast()
}
