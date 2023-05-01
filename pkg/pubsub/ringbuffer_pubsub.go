package pubsub

import (
	"github.com/tidwall/match"
	"sync"
)

/*

This was an experiment in implementing a lossy pubsub without using chans/callbacks that might slow down the
publishing goroutine.

Paused for now - the cond.Wait() requires the lock to be acquired(not even RLock), which makes the messages
queue up and be delivered to all subscribers sequentially(as in reader 1 gets all messages in a row, then on to reader 2).
This seems to be an effect of how the cond package handles the waiters: they're put on a list and only dequeued when they release
the lock, either by Unlock or cond.Wait - in this case that happens when all messages are read.

Dream scenario here would be a cond.Broadcast() + cond.Wait() that does not require locking of the resource/that can work with a
RLock, not a Lock.

*/

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
