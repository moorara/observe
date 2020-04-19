package observe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMeter(t *testing.T) {
	tests := []struct {
		name string
		opts MeterOptions
	}{
		{
			name: "WithName",
			opts: MeterOptions{
				Name: "my-service",
			},
		},
		{
			name: "WithSystemMetrics",
			opts: MeterOptions{
				Name:          "my-service",
				SystemMetrics: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meter, close, handler := NewMeter(tc.opts)
			defer close()

			assert.NotNil(t, meter)
			assert.NotNil(t, close)
			assert.NotNil(t, handler)
		})
	}
}
