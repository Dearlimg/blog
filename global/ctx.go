package global

import (
	"context"
	"time"
)

func DefaultContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
