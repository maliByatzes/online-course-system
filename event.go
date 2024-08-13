package ocs

import "context"

// define event type constraints

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// define structs here...

type EventService interface {
	PublishEvent(studentID int, event Event)
	Subscribe(ctx context.Context) (Subscription, error)
}

func NopEventService() EventService { return &nopEventService{} }

type nopEventService struct{}

func (*nopEventService) PublishEvent(studentID int, event Event) {}

func (*nopEventService) Subscribe(ctx context.Context) (Subscription, error) {
	panic("not implemented")
}

type Subscription interface {
	C() <-chan Event
	Close() error
}
