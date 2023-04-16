package pubsub_test

import (
	"sync"
	"testing"
	"time"

	"github.com/fnurk/geom/pkg/pubsub"
)

func TestPubSub_TestSimpleDelivery(t *testing.T) {
	ps := pubsub.New(1000)
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
			func(m *pubsub.Message) {
				numMsgs++
				counter++
				if numMsgs >= numPublishers {
					wg.Done()
				}
			}, func() {

			})
		subWg.Done()
	}

	subWg.Wait()

	for i := 0; i < numPublishers; i++ {
		go func(id int) {
			ps.Publish(pubsub.Message{
				Topic: "hello",
				Body:  "world"})
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

	counter := 0

	ps.Subscribe("hello", func(m *pubsub.Message) {
		counter++
	}, func() {
	})

	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})
	ps.Publish(pubsub.Message{Topic: "hello", Body: "msg1"})

	time.Sleep(1 * time.Millisecond)

	if counter != 9 {
		t.Errorf("expected 9, got %d", counter)
	}
}

func TestPubSub_NoSubs(t *testing.T) {
	ps := pubsub.New(1000)
	numPublishers := 1000

	for i := 0; i < numPublishers; i++ {
		ps.Publish(pubsub.Message{Topic: "hello", Body: "world"})
	}
}

func Benchmark_PubSub_NoSubs(b *testing.B) {
	ps := pubsub.New(1000)
	numPublishers := 1000

	for i := 0; i < b.N; i++ {
		for i := 0; i < numPublishers; i++ {
			ps.Publish(pubsub.Message{Topic: "hello", Body: "world"})
		}

	}
}

//Run 1000 publishers, one message each, to 1000 subscribers
func Benchmark_PubSub(b *testing.B) {
	ps := pubsub.New(100)

	for i := 0; i < 1000; i++ {
		ps.Subscribe("hello",
			func(m *pubsub.Message) {
				//noop
			}, func() {})

	}

	for i := 0; i < b.N; i++ {
		ps.Publish(pubsub.Message{Topic: "hello", Body: "world"})
	}

	ps.Shutdown()
}
