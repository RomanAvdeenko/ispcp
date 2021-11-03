package model

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Ping implements arping request parameters
type Ping struct {
	IP    net.IP
	Iface net.Interface
}

// Pong contains data about the host availability
type Pong struct {
	IpAddr   net.IP
	MACAddr  net.HardwareAddr
	Alive    bool
	Duration time.Duration
	Time     time.Time
}

type Pongs struct {
	sync.RWMutex
	pong []Pong
}

func NewPongs() *Pongs {
	return &Pongs{pong: make([]Pong, 0, 32)}
}

func (p *Pongs) Store(key uint32, val *Pong) {
	p.Lock()
	defer p.Unlock()

	p.pong = append(p.pong, *val)
}

func (p *Pongs) LoadAll() *[]Pong {
	return &p.pong
}

func (p *Pongs) Clear() {
	p.Lock()
	defer p.Unlock()

	p.pong = p.pong[:0]
}

// Pong human friendly view implementation
func (p *Pong) Human() string {
	return fmt.Sprintf("ip: %s,\talive: %s,\tmac: %s,\tdate: %s,\tduration: %s", p.IpAddr, p.IpAddr, p.MACAddr, p.Time.Format(time.Stamp), p.Duration)
}
