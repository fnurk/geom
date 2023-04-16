package pubsub_test

import (
	"sync"
	"testing"

	"github.com/fnurk/geom/pkg/pubsub"
)

func TestPubSub_TestSimpleDelivery(t *testing.T) {
	ps := pubsub.New()
	numSubs := 1000
	numPublishers := 1000

	subWg := sync.WaitGroup{}
	subWg.Add(numSubs)

	wg := sync.WaitGroup{}
	wg.Add(numSubs + numPublishers)

	for i := 0; i < numSubs; i++ {
		go func(id int) {
			s := ps.Subscribe("hello")
			subWg.Done()

			numMsgs := 0
			for numMsgs < numPublishers {
				<-s.MessageChan
				numMsgs++
			}
			s.Unsubscribe()

			wg.Done()
		}(i)
	}

	subWg.Wait()

	for i := 0; i < numPublishers; i++ {
		go func(id int) {
			ps.Publish("hello", pubsub.Message{Body: "world"})
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestPubSub_Unsubscribe(t *testing.T) {
	ps := pubsub.New()

	s1 := ps.Subscribe("hello")
	s2 := ps.Subscribe("hello")

	ps.Publish("hello", pubsub.Message{Body: "world"})

	<-s1.MessageChan
	<-s2.MessageChan

	s2.Unsubscribe()

	ps.Publish("hello", pubsub.Message{Body: "world"})

	<-s1.MessageChan
}

func TestPubSub_NoSubs(t *testing.T) {
	ps := pubsub.New()
	numPublishers := 1000

	wg := sync.WaitGroup{}
	wg.Add(numPublishers)

	for i := 0; i < numPublishers; i++ {
		go func(id int) {
			ps.Publish("hello", pubsub.Message{Body: "world"})
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func Benchmark_PubSub_NoSubs(b *testing.B) {
	ps := pubsub.New()
	numPublishers := 1000

	for i := 0; i < b.N; i++ {
		for i := 0; i < numPublishers; i++ {
			go func(id int) {
				ps.Publish("hello", pubsub.Message{Body: "world"})
			}(i)
		}

	}
}

//Run 1000 publishers, one message each, to 1000 subscribers
func Benchmark_PubSub(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ps := pubsub.New()
		numSubs := 1000
		numPublishers := 1000

		subWg := sync.WaitGroup{}
		subWg.Add(numSubs)

		wg := sync.WaitGroup{}
		wg.Add(numSubs + numPublishers)

		for i := 0; i < numSubs; i++ {
			go func(id int) {
				s := ps.Subscribe("hello")
				subWg.Done()

				numMsgs := 0
				for numMsgs < numPublishers {
					<-s.MessageChan
					numMsgs++
				}

				s.Unsubscribe()

				wg.Done()
			}(i)
		}

		subWg.Wait()

		for i := 0; i < numPublishers; i++ {
			go func(id int) {
				ps.Publish("hello", pubsub.Message{Body: "world"})
				wg.Done()
			}(i)
		}

		wg.Wait()
	}
}
