package model

import (
	"net"
	"time"
)

// Pong contains data about the host availability
type Pong struct {
	IpAddr   net.IP
	MacAddr  net.HardwareAddr
	Alive    bool
	RespTime time.Duration
}
