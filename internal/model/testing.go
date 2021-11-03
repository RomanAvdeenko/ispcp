package model

import (
	"net"
	"testing"
	"time"
)

func TestCheckHostRecord(t *testing.T) *Pong {
	t.Helper()

	ip := net.ParseIP("8.8.8.8")
	mac, _ := net.ParseMAC("b4:d5:bd:b8:c1:94")
	time := time.Duration(0)

	return &Pong{
		IpAddr:   ip,
		MACAddr:  mac,
		RespTime: time,
	}
}
