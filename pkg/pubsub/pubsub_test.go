package pubsub_test

import (
	"sync"
	"testing"

	"github.com/fnurk/geom/pkg/pubsub"
	"go.uber.org/goleak"
)

func TestChanPubSub_MessageCount(t *testing.T) {
	defer goleak.VerifyNone(t)

	ps := pubsub.NewChanPubsub()

	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		counter := 0
		ps.Subscribe("hello", func(m *pubsub.Message, s pubsub.Subscriber) {
			counter++
			if counter == 1000 {
				wg.Done()
			}
		}, func() {
		})
	}

	for i := 0; i < 1000; i++ {
		ps.Publish(&pubsub.Message{Topic: "hello", Body: "msg1"})
	}

	wg.Wait()

	ps.Shutdown()
}

func TestChanPubSub_Unsubscribe(t *testing.T) {
	defer goleak.VerifyNone(t)

	ps := pubsub.NewChanPubsub()

	counter := 0
	ps.Subscribe("hello", func(m *pubsub.Message, s pubsub.Subscriber) {
		counter++
		if counter == 5 {
			s.Unsubscribe()
		}
	}, func() {
	})

	for i := 0; i < 10; i++ {
		ps.Publish(&pubsub.Message{Topic: "hello", Body: "msg1"})
	}

	if counter != 5 {
		t.Errorf("expected 5, got %d", counter)
	}
}

func Benchmark_ChanPubSub_ExactMatch(b *testing.B) {
	ps := pubsub.NewChanPubsub()

	ps.Subscribe("hello",
		func(m *pubsub.Message, s pubsub.Subscriber) {
			//noop
		}, func() {})

	for i := 0; i < b.N; i++ {
		ps.Publish(&pubsub.Message{Topic: "hello", Body: "world"})
	}

	ps.Shutdown()
}

func Benchmark_ChanPubSub_Wildcard(b *testing.B) {
	ps := pubsub.NewChanPubsub()

	ps.Subscribe("*",
		func(m *pubsub.Message, s pubsub.Subscriber) {
			//noop
		}, func() {})

	for i := 0; i < b.N; i++ {
		ps.Publish(&pubsub.Message{Topic: "hello", Body: "world"})
	}

	ps.Shutdown()
}

func Benchmark_ChanPubSub_1000_Fanout(b *testing.B) {
	ps := pubsub.NewChanPubsub()

	for i := 0; i < 1000; i++ {
		ps.Subscribe("hello",
			func(m *pubsub.Message, s pubsub.Subscriber) {
				//noop
			}, func() {})
	}

	for i := 0; i < b.N; i++ {
		ps.Publish(&pubsub.Message{Topic: "hello", Body: "world"})
	}

	ps.Shutdown()
}
