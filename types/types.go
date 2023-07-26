package types

import "time"

type EventType string

var (
	EventTypeActionFinish EventType = "finish"
	EventTypeActionStart  EventType = "start"
	EventTypeOutput       EventType = "output"
)

type EventMsg struct {
	EventType
	Duration time.Duration
	Message  string
}
