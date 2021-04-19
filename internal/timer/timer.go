package timer

import (
	"time"
)

type RaftTimer interface {
	SetDeadline(currentTime time.Time)
	GetDeadline() time.Time
	SetIdleTimeout()
	GetIdleTimeout() time.Time
}
