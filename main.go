package main

import "github.com/moorara/observe/log"

func main() {
  logger := log.NewLogger(log.Options{
    Format:      log.JSON,
    Level:       "debug",
    Name:        "hello-world",
    Environment: "production",
    Region:      "us-east-1",
  })

  logger.Debug(
    "message", "Hello, World!",
    "context", map[string]interface{}{
      "retries": 4,
    },
  )
}
