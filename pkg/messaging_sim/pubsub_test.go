package messaging_sim

import (
	"sync"
	"testing"
	"time"
)

func TestPubSubCreation(t *testing.T) {
	ps := NewPubSub()
	
	if ps.subscribers == nil {
		t.Error("Subscribers map should be initialized")
	}
}

func TestPubSubSubscribe(t *testing.T) {
	ps := NewPubSub()
	
	// Subscribe to a topic
	ch := ps.Subscribe("test-topic")
	
	if ch == nil {
		t.Error("Subscribe should return a non-nil channel")
	}
	
	// Check that the channel is buffered
	if cap(ch) != 10 {
		t.Errorf("Expected channel buffer capacity of 10, got %d", cap(ch))
	}
	
	// Check that the subscription was added
	if len(ps.subscribers["test-topic"]) != 1 {
		t.Errorf("Expected 1 subscriber for test-topic, got %d", len(ps.subscribers["test-topic"]))
	}
}

func TestPubSubPublish(t *testing.T) {
	ps := NewPubSub()
	
	// Subscribe to a topic
	ch := ps.Subscribe("test-topic")
	
	// Publish a message
	payload := map[string]string{"key": "value"}
	ps.Publish("test-topic", payload)
	
	// Wait for the message
	select {
	case msg := <-ch:
		if msg.Topic != "test-topic" {
			t.Errorf("Expected topic test-topic, got %s", msg.Topic)
		}
		
		if p, ok := msg.Payload.(map[string]string); !ok {
			t.Errorf("Expected payload type map[string]string, got %T", msg.Payload)
		} else if p["key"] != "value" {
			t.Errorf("Expected payload[\"key\"] to be 'value', got %s", p["key"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timed out waiting for message")
	}
	
	// Publish to a topic with no subscribers (should not block or panic)
	ps.Publish("nonexistent-topic", "test")
}

func TestPubSubUnsubscribe(t *testing.T) {
	ps := NewPubSub()
	
	// Subscribe to a topic
	ch := ps.Subscribe("test-topic")
	
	// Unsubscribe
	ps.Unsubscribe("test-topic", ch)
	
	// Check that the subscription was removed
	if _, ok := ps.subscribers["test-topic"]; ok {
		t.Error("Expected test-topic to be removed from subscribers map")
	}
	
	// Unsubscribe from a topic that doesn't exist (should not panic)
	ps.Unsubscribe("nonexistent-topic", ch)
	
	// Subscribe multiple channels to the same topic
	ch1 := ps.Subscribe("multi-topic")
	ch2 := ps.Subscribe("multi-topic")
	
	// Check that both subscriptions were added
	if len(ps.subscribers["multi-topic"]) != 2 {
		t.Errorf("Expected 2 subscribers for multi-topic, got %d", len(ps.subscribers["multi-topic"]))
	}
	
	// Unsubscribe one channel
	ps.Unsubscribe("multi-topic", ch1)
	
	// Check that only one subscription remains
	if len(ps.subscribers["multi-topic"]) != 1 {
		t.Errorf("Expected 1 subscriber for multi-topic after unsubscribe, got %d", len(ps.subscribers["multi-topic"]))
	}
	
	// Unsubscribe the last channel
	ps.Unsubscribe("multi-topic", ch2)
	
	// Check that the topic was removed
	if _, ok := ps.subscribers["multi-topic"]; ok {
		t.Error("Expected multi-topic to be removed from subscribers map")
	}
}

func TestPubSubClose(t *testing.T) {
	ps := NewPubSub()
	
	// Subscribe to multiple topics
	ch1 := ps.Subscribe("topic1")
	ch2 := ps.Subscribe("topic2")
	
	// Close the PubSub
	ps.Close()
	
	// Check that all channels are closed
	select {
	case _, ok := <-ch1:
		if ok {
			t.Error("Expected ch1 to be closed")
		}
	default:
		t.Error("Expected ch1 to be closed and readable")
	}
	
	select {
	case _, ok := <-ch2:
		if ok {
			t.Error("Expected ch2 to be closed")
		}
	default:
		t.Error("Expected ch2 to be closed and readable")
	}
	
	// Check that subscribers map is empty
	if len(ps.subscribers) != 0 {
		t.Errorf("Expected subscribers map to be empty, got %d entries", len(ps.subscribers))
	}
}

func TestPubSubConcurrency(t *testing.T) {
	ps := NewPubSub()
	var wg sync.WaitGroup
	
	// Create multiple subscribers
	channels := make([]chan Message, 10)
	for i := 0; i < 10; i++ {
		channels[i] = ps.Subscribe("concurrent-topic")
	}
	
	// Publish messages concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ps.Publish("concurrent-topic", idx)
		}(i)
	}
	
	// Wait for all publishers
	wg.Wait()
	
	// Check that all subscribers received all messages
	receivedCounts := make([]int, 10)
	
	// Set a timeout for receiving messages
	timeout := time.After(500 * time.Millisecond)
	
	// Try to receive 5 messages for each subscriber
	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			select {
			case <-channels[i]:
				receivedCounts[i]++
			case <-timeout:
				// If we timeout, we'll just count what we've received so far
				j = 5 // Break out of the inner loop
			}
		}
	}
	
	// Check that each subscriber received all messages
	for i, count := range receivedCounts {
		if count != 5 {
			t.Errorf("Subscriber %d received %d messages, expected 5", i, count)
		}
	}
	
	// Test concurrent subscribe/unsubscribe
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			topic := "topic-" + string(rune('A'+idx))
			ch := ps.Subscribe(topic)
			ps.Publish(topic, idx)
			
			// Receive the message
			select {
			case msg := <-ch:
				if msg.Payload != idx {
					t.Errorf("Expected payload %d, got %v", idx, msg.Payload)
				}
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Timed out waiting for message on topic %s", topic)
			}
			
			ps.Unsubscribe(topic, ch)
		}(i)
	}
	
	wg.Wait()
}
