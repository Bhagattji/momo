package monitoring

import (
	"log"
)

// Simple monitoring/logging stub. Replace with metrics/otel as needed.

func Init() error {
	// initialize logging/metrics exporters here
	log.Println("monitoring: init (stub)")
	return nil
}

func Shutdown() error {
	log.Println("monitoring: shutdown (stub)")
	return nil
}
