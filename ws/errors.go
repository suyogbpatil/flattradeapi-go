package ws

import "fmt"

// ConnectionError reports a WebSocket dial, read, heartbeat, or write failure.
type ConnectionError struct {
	Op  string
	Err error
}

func (e *ConnectionError) Error() string { return fmt.Sprintf("websocket %s failed: %v", e.Op, e.Err) }
func (e *ConnectionError) Unwrap() error { return e.Err }

// BrokerLoginError reports a rejected or invalid broker login acknowledgement.
type BrokerLoginError struct {
	Status string
	Err    error
}

func (e *BrokerLoginError) Error() string {
	if e.Status != "" {
		return fmt.Sprintf("websocket broker login failed: status %q", e.Status)
	}
	return fmt.Sprintf("websocket broker login failed: %v", e.Err)
}
func (e *BrokerLoginError) Unwrap() error { return e.Err }

// SubscriptionError reports a rejected or failed subscription operation.
type SubscriptionError struct {
	Event   Event
	Restore bool
	Err     error
}

func (e *SubscriptionError) Error() string {
	action := "subscription"
	if e.Restore {
		action = "subscription restoration"
	}
	return fmt.Sprintf("websocket %s failed for event %q: %v", action, e.Event, e.Err)
}
func (e *SubscriptionError) Unwrap() error { return e.Err }

// MessageError reports a broker message that could not be decoded or handled.
type MessageError struct {
	Event Event
	Err   error
}

func (e *MessageError) Error() string {
	if e.Event != "" {
		return fmt.Sprintf("websocket message handling failed for event %q: %v", e.Event, e.Err)
	}
	return fmt.Sprintf("websocket message decoding failed: %v", e.Err)
}
func (e *MessageError) Unwrap() error { return e.Err }
