package domain

import "time"

type Message struct {
	Data      string
	Timestamp time.Time
}

type Broker interface {
	Put(queueName, msg string) error
	Get(queueName string, timeout time.Duration) (Message, error)
}
