package messaging_sim

import (
	"sync"
)

// Message represents a message in the pub/sub system
type Message struct {
	Topic   string
	Payload interface{}
}

// PubSub provides a simple publish-subscribe mechanism
type PubSub struct {
	subscribers map[string][]chan Message
	mutex       sync.RWMutex
}

// NewPubSub creates a new pub/sub system
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]chan Message),
	}
}

// Subscribe creates a subscription to a topic and returns a channel for receiving messages
func (ps *PubSub) Subscribe(topic string) chan Message {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Create a buffered channel to prevent blocking publishers
	ch := make(chan Message, 10)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

// Unsubscribe removes a subscription from a topic
func (ps *PubSub) Unsubscribe(topic string, ch chan Message) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	subs, ok := ps.subscribers[topic]
	if !ok {
		return
	}

	// Find and remove the channel
	for i, sub := range subs {
		if sub == ch {
			// Remove the channel from the slice
			ps.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	// If no more subscribers for this topic, remove the topic
	if len(ps.subscribers[topic]) == 0 {
		delete(ps.subscribers, topic)
	}
}

// Publish sends a message to all subscribers of a topic
func (ps *PubSub) Publish(topic string, payload interface{}) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	// If no subscribers, just return
	subs, ok := ps.subscribers[topic]
	if !ok {
		return
	}

	// Create the message
	msg := Message{
		Topic:   topic,
		Payload: payload,
	}

	// Send to all subscribers (non-blocking)
	for _, ch := range subs {
		select {
		case ch <- msg:
			// Message sent successfully
		default:
			// Channel is full, skip this subscriber
		}
	}
}

// Close closes all subscription channels
func (ps *PubSub) Close() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Close all channels
	for topic, subs := range ps.subscribers {
		for _, ch := range subs {
			close(ch)
		}
		delete(ps.subscribers, topic)
	}
}
