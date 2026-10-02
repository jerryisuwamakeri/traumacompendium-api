package fakes

import (
	"context"
	"sync"
)

// Published is one message recorded by Publisher.Publish.
type Published struct {
	Channel string
	Message any
}

// Publisher is an in-memory safereachservice.Publisher.
type Publisher struct {
	mu        sync.Mutex
	Published []Published
}

func (f *Publisher) Publish(_ context.Context, channel string, message any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Published = append(f.Published, Published{Channel: channel, Message: message})
	return nil
}

func (f *Publisher) Last() Published {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Published[len(f.Published)-1]
}
