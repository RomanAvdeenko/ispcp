package model

import (
	"net"
	"time"
)

// CheckHostRecord contains data about the host's network availability
type CheckHostRecord struct {
	IP    net.IP
	MAC   net.HardwareAddr
	Alive bool
	Time  time.Time
}
