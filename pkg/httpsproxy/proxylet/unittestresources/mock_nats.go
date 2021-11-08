package unittestresources

import "github.com/theotw/natssync/pkg/httpsproxy/nats"

type mockNats struct {
	lastError error
	flushError error
	publishError error
	subscribeError error
	queues map[string] []nats.Msg
}

func (m mockNats) Subscribe(subj string, cb nats.MsgHandler) (nats.NatsSubscriptionInterface, error) {

	if msgs, ok := m.queues[subj]; ok {
		for _,msg := range msgs {
			cb(&msg)
		}
	}

	return nil, m.subscribeError
}

func (m mockNats) Publish(subj string, data []byte) error {
	if _, ok := m.queues[subj]; !ok {
		m.queues[subj] = make([]nats.Msg,0)
	}
	m.queues[subj] = append(m.queues[subj], nats.Msg{Data: data})
	return m.publishError
}

func (m mockNats) LastError() error {
	return m.lastError
}

func (m mockNats) Flush() error {
	return m.flushError
}

func (m mockNats) SubscribeSync(reply string) (nats.NatsSubscriptionInterface, error) {
	panic("implement me")
}

func (m mockNats) PublishRequest(subj string, reply string, data []byte) error {
	panic("implement me")
}


// ================================================
type MockNatsInput struct {
	lastError error
	flushError error
	publishError error
	subscribeError error
	queues map[string][]nats.Msg
}

func NewDefaultMockNatsInput() MockNatsInput {
	return MockNatsInput{
		lastError: nil,
		flushError: nil,
		publishError: nil,
		queues: make(map[string][]nats.Msg),
	}
}

func NewMockNats(input MockNatsInput) *mockNats {
	return &mockNats{
		lastError:  input.lastError,
		flushError: input.flushError,
		publishError: input.publishError,
		queues: input.queues,
	}
}
