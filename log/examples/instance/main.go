package main

import "github.com/moorara/observe/log"

func main() {
	logger := log.NewLogger(log.Options{
		Name:        "service",
		Environment: "production",
		Region:      "us-east-1",
		Level:       "debug",
		Format:      log.JSON,
	})

	logger = logger.With(
		"version", "0.1.0",
		"revision", "abcdef",
	)

	logger.DebugKV(
		"message", "Hello, World!",
		"requestId", "2222-bbbb",
	)
}
