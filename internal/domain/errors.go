package domain

import "errors"

var (
	ErrQueueFull       = errors.New("queue is full")
	ErrQueueLimit      = errors.New("maximum queues limit reached")
	ErrMessageNotFound = errors.New("message not found")
)
