package session

import (
	"fmt"
	"sync"
	"time"
)

// EventType represents the type of event in the pub/sub system.
type EventType string

const (
	EventTypeFileChanged   EventType = "file_changed"
	EventTypeSessionCreated EventType = "session_created"
	EventTypeSessionClosed EventType = "session_closed"
	EventTypeUserJoined    EventType = "user_joined"
	EventTypeUserLeft     EventType = "user_left"
	EventTypeOperation     EventType = "operation"
)

// PubSubEvent represents an event in the pub/sub system.
type PubSubEvent struct {
	Type      EventType
	Channel   string // File path or session ID
	Timestamp int64
	Data      interface{}
	Metadata  map[string]interface{}
}

// Subscription represents a subscription to a channel/file.
type Subscription struct {
	ID         string
	Channel    string
	Subscriber string // User ID or client ID
	EventChan  chan *PubSubEvent
	filter     func(*PubSubEvent) bool
	closeCh    chan struct{}
	closeOnce  sync.Once // Prevent double-close
	mu         sync.RWMutex
}

// PubSub provides Redis-like pub/sub functionality for file events.
type PubSub struct {
	mu            sync.RWMutex
	subscriptions map[string][]*Subscription // channel -> subscriptions
	channels      map[string]chan *PubSubEvent   // channel -> event broadcast
	closeCh       chan struct{}
	closed        bool
}

// NewPubSub creates a new pub/sub system.
func NewPubSub() *PubSub {
	return &PubSub{
		subscriptions: make(map[string][]*Subscription),
		channels:       make(map[string]chan *PubSubEvent),
		closeCh:        make(chan struct{}),
	}
}

// Subscribe subscribes to events for a specific channel/file.
// The subscription persists even if no session exists for the file.
//
// Parameters:
//   - channel: the file path or channel to subscribe to
//   - subscriber: unique identifier for the subscriber
//   - filter: optional filter function to receive only specific events
//
// Returns:
//   - a subscription object with an event channel
//
// Example:
//
//	sub := pubsub.Subscribe("/path/to/file.txt", "user1", nil)
//	for event := range sub.EventChan {
//	    fmt.Printf("Event: %+v\n", event)
//	}
func (ps *PubSub) Subscribe(channel, subscriber string, filter func(*PubSubEvent) bool) (*Subscription, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil, fmt.Errorf("pubsub system is closed")
	}

	// Ensure channel exists
	if _, ok := ps.channels[channel]; !ok {
		ps.channels[channel] = make(chan *PubSubEvent, 100)
	}

	sub := &Subscription{
		ID:         fmt.Sprintf("%s-%s-%d", subscriber, channel, time.Now().UnixNano()),
		Channel:    channel,
		Subscriber: subscriber,
		EventChan:  make(chan *PubSubEvent, 100),
		filter:     filter,
		closeCh:    make(chan struct{}),
	}

	ps.subscriptions[channel] = append(ps.subscriptions[channel], sub)

	// Start forwarding events to this subscriber
	go ps.forwardEvents(sub)

	return sub, nil
}

// forwardEvents forwards events from the channel to the subscription.
func (ps *PubSub) forwardEvents(sub *Subscription) {
	for {
		select {
		case <-ps.closeCh:
			return
		case <-sub.closeCh:
			return
		case event, ok := <-ps.channels[sub.Channel]:
			if !ok {
				return
			}

			// Apply filter if provided
			if sub.filter != nil && !sub.filter(event) {
				continue
			}

			select {
			case sub.EventChan <- event:
			case <-sub.closeCh:
				return
			case <-ps.closeCh:
				return
			}
			}
	}
}

// Publish publishes an event to a channel.
// All subscribers to the channel will receive the event.
//
// Parameters:
//   - channel: the file path or channel
//   - event: the event to publish
//
// Returns:
//   - error if the pubsub system is closed
//
// Example:
//
//	pubsub.Publish("/path/to/file.txt", &PubSubEvent{
//	    Type: EventTypeFileChanged,
//	    Channel: "/path/to/file.txt",
//	    Data: "new content",
//	})
func (ps *PubSub) Publish(channel string, event *PubSubEvent) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return fmt.Errorf("pubsub system is closed")
	}

	event.Channel = channel
	event.Timestamp = time.Now().Unix()

	// Get or create channel
	ch, ok := ps.channels[channel]
	if !ok {
		ch = make(chan *PubSubEvent, 100)
		ps.channels[channel] = ch
	}

	// Publish to channel (non-blocking)
	select {
	case ch <- event:
	default:
		// Channel full or no subscribers, drop event
	}

	return nil
}

// Unsubscribe cancels a subscription.
func (ps *PubSub) Unsubscribe(sub *Subscription) error {
	sub.mu.Lock()
	defer sub.mu.Unlock()

	// Check if already closed (read without select to avoid blocking)
	select {
	case <-sub.closeCh:
		return fmt.Errorf("already unsubscribed")
	default:
		// Not closed yet, use sync.Once to close
		sub.closeOnce.Do(func() {
			close(sub.closeCh)
		})
	}

	// Remove from subscriptions map
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subs := ps.subscriptions[sub.Channel]
	for i, s := range subs {
		if s.ID == sub.ID {
			// Remove from slice
			ps.subscriptions[sub.Channel] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	return nil
}

// GetSubscribers returns all subscribers for a channel.
func (ps *PubSub) GetSubscribers(channel string) []*Subscription {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subs := ps.subscriptions[channel]
	result := make([]*Subscription, len(subs))
	copy(result, subs)
	return result
}

// ListChannels returns all active channels.
func (ps *PubSub) ListChannels() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	channels := make([]string, 0, len(ps.channels))
	for ch := range ps.channels {
		channels = append(channels, ch)
	}
	return channels
}

// Close closes the pub/sub system.
func (ps *PubSub) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil
	}

	ps.closed = true

	// Step 1: Signal all goroutines to stop by closing subscription closeChs
	for _, subs := range ps.subscriptions {
		for _, sub := range subs {
			select {
			case <-sub.closeCh:
				// Already closed
			default:
				close(sub.closeCh)
			}
		}
	}

	// Step 2: Close the pubsub closeCh to signal any waiting operations
	close(ps.closeCh)

	// Step 3: Close all broadcast channels (after goroutines stopped)
	for _, ch := range ps.channels {
		close(ch)
	}

	// Step 4: Close all subscription EventChans (finally)
	for _, subs := range ps.subscriptions {
		for _, sub := range subs {
			select {
			case <-sub.EventChan:
				// Drain channel
			default:
			}
			close(sub.EventChan)
		}
	}

	ps.subscriptions = make(map[string][]*Subscription)
	ps.channels = make(map[string]chan *PubSubEvent)

	return nil
}

// GetChannelInfo returns information about a channel.
func (ps *PubSub) GetChannelInfo(channel string) *ChannelInfo {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subs := ps.subscriptions[channel]
	return &ChannelInfo{
		Channel:      channel,
		SubscriberCount: len(subs),
		Subscribers:  subs,
	}
}

// ChannelInfo contains information about a channel.
type ChannelInfo struct {
	Channel         string
	SubscriberCount int
	Subscribers     []*Subscription
}
