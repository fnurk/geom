package pubsub_test

import (
	"sync"
	"testing"

	"github.com/fnurk/geom/pkg/pubsub"
)

func TestPubSub_TestSimpleDelivery(t *testing.T) {
	ps := pubsub.New(100)
	numSubs := 100
	numPublishers := 100

	subWg := sync.WaitGroup{}
	subWg.Add(numSubs)

	wg := sync.WaitGroup{}
	wg.Add(numSubs + numPublishers)

	counter := 0

	for i := 0; i < numSubs; i++ {
		numMsgs := 0
		ps.Subscribe("hello",
			func(s *pubsub.Subscriber, m pubsub.Message) {
				numMsgs++
				counter++
				if numMsgs >= numPublishers {
					wg.Done()
					s.Active = false
				}
			}, func() {

			})
		subWg.Done()
	}

	subWg.Wait()

	for i := 0; i < numPublishers; i++ {
		go func(id int) {
			ps.Publish("hello", pubsub.Message{Body: "world"})
			wg.Done()
		}(i)
	}

	wg.Wait()

	if counter != (numSubs * numPublishers) {
		t.Errorf("FAILEDD, got %d", counter)
	}
}

func TestPubSub_Unsubscribe(t *testing.T) {
	ps := pubsub.New(1000)

	ps.Subscribe("hello", func(s *pubsub.Subscriber, m pubsub.Message) {
		t.Logf("S1: Got message %s", m.Body)
	}, func() {
		t.Logf("s2 forcefully shutdown")
	})

	ps.Subscribe("hello", func(s *pubsub.Subscriber, m pubsub.Message) {
		t.Logf("S2: Got message %s", m.Body)
	}, func() {
		t.Logf("s2 forcefully shutdown")
	})

	go func() {
		ps.Publish("hello", pubsub.Message{Body: "msg1"})
		ps.Publish("hello", pubsub.Message{Body: "msg2"})
		ps.Publish("hello", pubsub.Message{Body: "msg3"})
		ps.Publish("hello", pubsub.Message{Body: "msg4"})
		ps.Publish("hello", pubsub.Message{Body: "msg5"})
		ps.Publish("hello", pubsub.Message{Body: "msg7"})
		ps.Publish("hello", pubsub.Message{Body: "msg8"})
		ps.Publish("hello", pubsub.Message{Body: "msg9"})
	}()
}

func TestPubSub_NoSubs(t *testing.T) {
	ps := pubsub.New(1000)
	numPublishers := 1000

	for i := 0; i < numPublishers; i++ {
		ps.Publish("hello", pubsub.Message{Body: "world"})
	}
}

func Benchmark_PubSub_NoSubs(b *testing.B) {
	ps := pubsub.New(1000)
	numPublishers := 1000

	for i := 0; i < b.N; i++ {
		for i := 0; i < numPublishers; i++ {
			ps.Publish("hello", pubsub.Message{Body: "world"})
		}

	}
}

//Run 1000 publishers, one message each, to 1000 subscribers
func Benchmark_PubSub(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ps := pubsub.New(10000)
		numSubs := 1000
		numPublishers := 1000

		subWg := sync.WaitGroup{}
		subWg.Add(numSubs)

		wg := sync.WaitGroup{}
		wg.Add(numSubs + numPublishers)

		for i := 0; i < numSubs; i++ {
			numMsgs := 0
			ps.Subscribe("hello",
				func(s *pubsub.Subscriber, m pubsub.Message) {
					numMsgs++
					if numMsgs >= numPublishers {
						s.Active = false
						wg.Done()
					}
				}, func() { wg.Done() })
			subWg.Done()
		}

		subWg.Wait()

		for i := 0; i < numPublishers; i++ {
			go func(id int) {
				ps.Publish("hello", pubsub.Message{Body: "world"})
				wg.Done()
			}(i)
		}

		ps.Shutdown()
		wg.Wait()
	}
}
