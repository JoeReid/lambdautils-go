package lambdautils

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	errDeadlineNotSet      = errors.New("lambdautils: context deadline not set")
	errGracePeriodNegative = errors.New("lambdautils: grace period cannot be negative")
	errGracePeriodTooLarge = errors.New("lambdautils: lambda times out before grace period ends")
)

var _ lambda.Handler = (*Graceful)(nil)

type Graceful struct {
	GracePeriod time.Duration
	Handler     lambda.Handler
}

func (g *Graceful) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	if g.GracePeriod < 0 {
		return []byte("null"), errGracePeriodNegative
	}

	lambdaDeadline, ok := ctx.Deadline()
	if !ok {
		return []byte("null"), errDeadlineNotSet
	}

	// Ensure that we have enough time to actually apply the grace period
	// before the lambda deadline is reached.
	// Otherwise, return an error rather than invoking the handler.
	newDeadline := lambdaDeadline.Add(-g.GracePeriod)
	if newDeadline.Before(time.Now()) {
		return []byte("null"), errGracePeriodTooLarge
	}

	ctx, cancel := context.WithDeadline(ctx, newDeadline)
	defer cancel()

	return g.Handler.Invoke(ctx, payload)
}

func WithGracefulShutdown(handler interface{}, gracePeriod time.Duration) *Graceful {
	return &Graceful{
		GracePeriod: gracePeriod,
		Handler:     lambda.NewHandler(handler),
	}
}
