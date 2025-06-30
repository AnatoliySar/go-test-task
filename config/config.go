package config

import "flag"

type Config struct {
	Port           int
	DefaultTimeout int
	MaxQueues      int
	MaxMessages    int
}

func NewConfig(port, defaultTimeout, maxQueues, maxMessages int) *Config {
	return &Config{
		Port:           port,
		DefaultTimeout: defaultTimeout,
		MaxQueues:      maxQueues,
		MaxMessages:    maxMessages,
	}
}

func ParseFlags() *Config {
	port := flag.Int("port", 8081, "port to listen on")
	defaultTimeout := flag.Int("timeout", 0, "default timeout in seconds (0 - no timeout)")
	maxQueues := flag.Int("max-queues", 100, "maximum number of queues")
	maxMessages := flag.Int("max-messages", 1000, "maximum messages per queue")
	flag.Parse()

	return NewConfig(*port, *defaultTimeout, *maxQueues, *maxMessages)
}
