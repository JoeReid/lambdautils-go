package lambdautils

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var _ lambda.Handler = (*Graceful)(nil)

type Graceful struct {
	GracePeriod time.Duration
	Handler     lambda.Handler
}

func (g *Graceful) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	lambdaDeadline, ok := ctx.Deadline()
	if !ok {
		return nil, errors.New("lambdautils: context deadline not set")
	}

	ctx, cancel := context.WithDeadline(ctx, lambdaDeadline.Add(-g.GracePeriod))
	defer cancel()

	return g.Handler.Invoke(ctx, payload)
}

func WithGracefulShutdown(handler interface{}, gracePeriod time.Duration) *Graceful {
	return &Graceful{
		GracePeriod: gracePeriod,
		Handler:     lambda.NewHandler(handler),
	}
}
