package unittestresources

import (
	"fmt"
	"sync"
	"time"

	"github.com/theotw/natssync/pkg/httpsproxy/nats"
)

type mockSubscription struct {
	autoUnsubscribeError error
	unsubscribeError     error
	nextMsgError         error
	Queue                []nats.Msg
	counter              int
}

func (m *mockSubscription) AutoUnsubscribe(max int) error {
	return m.autoUnsubscribeError
}

func (m *mockSubscription) NextMsg(duration time.Duration) (*nats.Msg, error) {
	if m.counter >= len(m.Queue) {
		for {
			if m.counter < len(m.Queue) {
				break
			}
		}
	}

	message := m.Queue[m.counter]
	m.counter = m.counter + 1
	return &message, m.nextMsgError
}

func (m *mockSubscription) Unsubscribe() error {
	return m.unsubscribeError
}

type mockNats struct {
	lock           sync.Mutex
	lastError      error
	flushError     error
	publishError   error
	subscribeError error
	Queues         map[string]*mockSubscription
}

func (m *mockNats) Subscribe(subj string, cb nats.MsgHandler) (nats.NatsSubscriptionInterface, error) {
	var subscription *mockSubscription
	var ok bool
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()
		subscription, ok = m.Queues[subj]
	}()

	if ok {
		for _, msg := range subscription.Queue {
			subscription.counter++
			cb(&msg)
		}
	}

	return nil, m.subscribeError
}
func (m *mockNats) QueueSubscribe(subj, queue string, cb nats.MsgHandler) (nats.NatsSubscriptionInterface, error) {
	var subscription *mockSubscription
	var ok bool
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()
		subscription, ok = m.Queues[subj]
	}()

	if ok {
		for _, msg := range subscription.Queue {
			subscription.counter++
			cb(&msg)
		}
	}

	return nil, m.subscribeError
}

func (m *mockNats) Publish(subj string, data []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.Queues[subj]; !ok {
		m.Queues[subj] = &mockSubscription{Queue: make([]nats.Msg, 0)}
	}
	msg := nats.Msg{Data: data}
	m.Queues[subj].Queue = append(m.Queues[subj].Queue, msg)
	return m.publishError
}

func (m *mockNats) LastError() error {
	return m.lastError
}

func (m *mockNats) Flush() error {
	return m.flushError
}

func (m *mockNats) SubscribeSync(subj string) (nats.NatsSubscriptionInterface, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.Queues[subj]; !ok {
		m.Queues[subj] = &mockSubscription{Queue: make([]nats.Msg, 0)}
	}

	return m.Queues[subj], nil
}

func (m *mockNats) PublishRequest(subj string, reply string, data []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Queues[subj] = &mockSubscription{Queue: []nats.Msg{{Data: data, Reply: reply}}}

	return nil
}

func (m *mockNats) PrintDebug() {
	for subject, subscription := range m.Queues {
		fmt.Printf("subject: %s count: %v : %v\n", subject, len(subscription.Queue), subscription.Queue)
	}
	fmt.Println()
}

// ================================================
type MockNatsInput struct {
	lastError      error
	flushError     error
	publishError   error
	subscribeError error
	queues         map[string]*mockSubscription
}

func NewDefaultMockNatsInput() MockNatsInput {
	return MockNatsInput{
		lastError:    nil,
		flushError:   nil,
		publishError: nil,
		queues:       make(map[string]*mockSubscription),
	}
}

func NewMockNats(input MockNatsInput) *mockNats {
	return &mockNats{
		lastError:    input.lastError,
		flushError:   input.flushError,
		publishError: input.publishError,
		Queues:       input.queues,
	}
}
