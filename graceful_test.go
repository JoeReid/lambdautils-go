package lambdautils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraceful_Invoke_deadlines(t *testing.T) {
	const timeErrorTollerance = time.Millisecond

	var tests = []struct {
		name          string
		lambdaTimeout time.Duration
		gracePeriod   time.Duration
		expectCalled  bool
		expectError   error
	}{
		{
			name:          "happy path",
			lambdaTimeout: 3 * time.Second,
			gracePeriod:   time.Second,
			expectCalled:  true,
			expectError:   nil,
		},
		{
			// This is a bit of a weird case, but it's a valid one.
			name:          "grace period zero",
			lambdaTimeout: 3 * time.Second,
			gracePeriod:   0,
			expectCalled:  true,
			expectError:   nil,
		},
		{
			name:          "grace period equal timeout",
			lambdaTimeout: 3 * time.Second,
			gracePeriod:   3 * time.Second,
			expectCalled:  false,
			expectError:   errGracePeriodTooLarge,
		},
		{
			name:          "grace period greater than timeout",
			lambdaTimeout: 3 * time.Second,
			gracePeriod:   4 * time.Second,
			expectCalled:  false,
			expectError:   errGracePeriodTooLarge,
		},
		{
			name:          "grace period negative",
			lambdaTimeout: 3 * time.Second,
			gracePeriod:   -1 * time.Second,
			expectCalled:  false,
			expectError:   errGracePeriodNegative,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				called = false
				now    = time.Now()
			)

			graceful := WithGracefulShutdown(func(ctx context.Context) error {
				called = true

				lambdaDeadline, ok := ctx.Deadline()
				require.True(t, ok)
				assert.WithinDuration(t, now.Add(tt.lambdaTimeout-tt.gracePeriod), lambdaDeadline, timeErrorTollerance)
				return nil
			}, tt.gracePeriod)

			ctx, cancel := context.WithTimeout(context.Background(), tt.lambdaTimeout)
			defer cancel()

			result, err := graceful.Invoke(ctx, nil)
			assert.Equal(t, tt.expectCalled, called)
			assert.ErrorIs(t, err, tt.expectError)
			assert.Equal(t, []byte("null"), result)
		})
	}
}
