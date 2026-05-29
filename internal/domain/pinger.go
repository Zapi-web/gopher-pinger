package domain

import (
	"context"
	"time"
)

type ActiveProcess struct {
	Cancel       context.CancelFunc
	IntervalChan chan time.Duration
}
