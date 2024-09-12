package mock

import (
	"context"

	"github.com/maliByatzes/ocs"
)

var _ ocs.EventService = (*EventService)(nil)

type EventService struct {
	PublishEventFn func(studentID int, event ocs.Event)
	SubscribeFn    func(ctx context.Context) (ocs.Subscription, error)
}

func (s *EventService) PublishEvent(studentID int, event ocs.Event) {
	s.PublishEvent(studentID, event)
}

func (s *EventService) Subscribe(ctx context.Context) (ocs.Subscription, error) {
	return s.Subscribe(ctx)
}

type Subscription struct {
	CloseFn func() error
	CFn     func() <-chan ocs.Event
}

func (s *Subscription) Close() error {
	return s.CloseFn()
}

func (s *Subscription) C() <-chan ocs.Event {
	return s.CFn()
}
