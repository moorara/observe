package observe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTracer(t *testing.T) {
	tests := []struct {
		name string
		opts TracerOptions
	}{
		{
			name: "WithAgent",
			opts: TracerOptions{
				Name: "my-service",
				Tags: [][2]string{
					{"version", "0.1.0"},
				},
				AgentEndpoint: "localhost:6831",
			},
		},
		{
			name: "WithCollector",
			opts: TracerOptions{
				Name: "my-service",
				Tags: [][2]string{
					{"version", "0.1.0"},
				},
				CollectorEndpoint: "http://localhost:14268/api/traces",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tracer, close := NewTracer(tc.opts)
			defer close()

			assert.NotNil(t, tracer)
			assert.NotNil(t, close)
		})
	}
}
