package rconfig

import (
	"time"
)

const PollDuration = 100 * time.Millisecond
const IdleTimeout = 1200 * time.Millisecond
const MaxPacketDelayMs = 400
const MinPacketDelayMs = 110
