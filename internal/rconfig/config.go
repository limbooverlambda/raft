package rconfig

import (
	"time"
)

const PollDuration = 100 * time.Millisecond
const IdleTimeout = 1200 * time.Millisecond
const MaxPacketDelayMs = 400
const MinPacketDelayMs = 110

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
