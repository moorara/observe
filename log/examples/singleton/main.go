package main

import "github.com/moorara/observe/log"

func main() {
	log.SetOptions(log.Options{
		Name:        "service",
		Environment: "production",
		Region:      "us-east-1",
	})

	log.Infof("Hello, %s!", "World")
}
