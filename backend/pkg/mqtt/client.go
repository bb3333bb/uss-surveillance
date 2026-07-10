package mqtt

import (
	"log"
	"sync"
	"time"
)

// MessageHandler handles incoming subscription events.
type MessageHandler func(topic string, payload string)

// Client defines the publisher-subscriber interface.
type Client interface {
	Publish(topic string, payload string) error
	Subscribe(topic string, handler MessageHandler) error
}

// MockClient models an in-memory MQTT message broker client.
type MockClient struct {
	mu          sync.RWMutex
	subscribers map[string][]MessageHandler
}

// NewMockClient creates a local message broker instance.
func NewMockClient() *MockClient {
	return &MockClient{
		subscribers: make(map[string][]MessageHandler),
	}
}

// DefaultClient is the package-level shared broker instance.
var DefaultClient = NewMockClient()

// Publish dispatches payload strings to matching subscribers.
func (c *MockClient) Publish(topic string, payload string) error {
	c.mu.RLock()
	handlers, exists := c.subscribers[topic]
	c.mu.RUnlock()

	if exists {
		for _, handler := range handlers {
			go handler(topic, payload)
		}
	}
	return nil
}

// Subscribe registers callback handlers to a topic.
func (c *MockClient) Subscribe(topic string, handler MessageHandler) error {
	c.mu.Lock()
	c.subscribers[topic] = append(c.subscribers[topic], handler)
	c.mu.Unlock()
	return nil
}

// StartDroneHubSimulator spawns a background routine modeling mechanical hub door steps.
func StartDroneHubSimulator(client Client) {
	log.Println("Starting background mechanical Drone Hub simulator...")
	err := client.Subscribe("drone/hub/command", func(topic string, payload string) {
		log.Printf("[Simulator Hub Event] Command received: %s", payload)
		if payload == "open_doors" {
			_ = client.Publish("drone/hub/status", "doors_opening")
			time.Sleep(1 * time.Second)
			_ = client.Publish("drone/hub/status", "doors_open")
			log.Println("[Simulator Hub Event] Status published: doors_open")
		} else if payload == "takeoff" {
			time.Sleep(1 * time.Second)
			_ = client.Publish("drone/hub/status", "takeoff_completed")
			log.Println("[Simulator Hub Event] Status published: takeoff_completed")
		}
	})
	if err != nil {
		log.Printf("Failed to start Drone Hub simulator subscription: %v", err)
	}
}

// RunLaunchSequence orchestrates the takeoff sequence with mechanical interlocks over MQTT topics.
func RunLaunchSequence(client Client) (bool, string) {
	statusChan := make(chan string, 10)

	err := client.Subscribe("drone/hub/status", func(topic string, payload string) {
		statusChan <- payload
	})
	if err != nil {
		return false, "Failed to subscribe to Drone Hub statuses"
	}

	// 1. Publish command to open mechanical cover doors
	log.Println("[Orchestrator] Dispatching door open command...")
	_ = client.Publish("drone/hub/command", "open_doors")

	// 2. Await Doors Confirmed Open status event (max 5 second timeout)
	doorsOpened := false
	timeout := time.After(5 * time.Second)

	for !doorsOpened {
		select {
		case status := <-statusChan:
			if status == "doors_open" {
				doorsOpened = true
			}
		case <-timeout:
			return false, "Timeout waiting for mechanical hub doors to open"
		}
	}

	// 3. Issue takeoff commands once safe interlock is verified
	log.Println("[Orchestrator] Interlock verified. Dispatching takeoff sequence command...")
	_ = client.Publish("drone/hub/command", "takeoff")

	// 4. Await takeoff ignition confirmation
	takeoffDone := false
	timeout = time.After(5 * time.Second)

	for !takeoffDone {
		select {
		case status := <-statusChan:
			if status == "takeoff_completed" {
				takeoffDone = true
			}
		case <-timeout:
			return false, "Timeout waiting for takeoff ignition confirmation"
		}
	}

	return true, "Drone takeoff sequence completed successfully"
}
