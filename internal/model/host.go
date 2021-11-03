package model

import (
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
	RespTime time.Duration
}

type Pongs struct {
	sync.RWMutex
	pong map[uint32]Pong
}

func NewPongs() *Pongs {
	return &Pongs{pong: make(map[uint32]Pong)}
}

func (p *Pongs) Store(key uint32, val *Pong) {
	p.Lock()
	defer p.Unlock()

	p.pong[key] = *val
}

func (p *Pongs) LoadAll() *map[uint32]Pong {
	res := make(map[uint32]Pong)
	p.RLock()
	defer p.RUnlock()

	for k, v := range p.pong {
		res[k] = v
	}
	return &res
}
