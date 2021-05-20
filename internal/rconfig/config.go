package rconfig

import (
	"time"
)

const PollDuration = 100 * time.Millisecond
const IdleTimeout = 8000 * time.Millisecond
const MaxPacketDelayMs = 10000
const MinPacketDelayMs = 4000

type MemberConfig struct {
	ID   string
	IP   string
	Port string
}

type Config struct {
	ServerID     string
	ServerPort   string
	ServerHost   string
	MemberConfig []MemberConfig
}
