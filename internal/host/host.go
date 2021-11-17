package host

import (
	"net"
	"os"
	"time"

	myslice "github.com/RomanAvdeenko/utils/slice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const estimatedIfaceNumber = 4

// Host implements unix host...
type Host struct {
	ProcessedIfaces   []net.Interface
	excludeIfaceNames []string
	excludeNetIPs     []string
}

func NewHost() *Host {
	return &Host{}
}

func (h *Host) SetExcludeIfaceNames(val []string) {
	if val != nil {
		h.excludeIfaceNames = val
	}
}

func (h *Host) SetExcludeNetworkIPs(val []string) {
	if val != nil {
		h.excludeNetIPs = val
	}
}

func (h *Host) Configure() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
	h.setProcessedIfaces()
}

func (h *Host) setProcessedIfaces() {
	ifaces, _ := net.Interfaces()
	processedInterfaces := make([]net.Interface, 0, estimatedIfaceNumber)
	for _, iface := range ifaces {
		// skip down interface & check next intf
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// skip loopback & check next intf
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// skip exclude interfaces from configuration
		if myslice.IsMatchesValue(h.excludeIfaceNames, iface.Name) {
			continue
		}
		processedInterfaces = append(processedInterfaces, iface)
	}
	h.ProcessedIfaces = processedInterfaces
}

func (h *Host) GetIfaceAddrs(iface net.Interface) (ipNets []string, err error) {
	res := make([]string, 0, estimatedIfaceNumber*2)
	addrs, err := iface.Addrs()
	if err != nil {
		log.Error().Msg("GetIfaceAddrs()")
		return
	}
	for _, addr := range addrs {
		ifaceIP, ok := addr.(*net.IPNet)
		if !ok {
			return
		}
		ip := ifaceIP.IP
		if ip == nil || ip.IsLoopback() {
			continue
		}
		// convert IP IPv4 address to 4-byte
		ip = ip.To4()
		if ip == nil {
			continue // not an ipv4 address
		}
		_, ifaceIPNet, err := net.ParseCIDR(ifaceIP.String())
		if err != nil {
			continue
		}
		// skip exclude network IP addresses from configuration
		if myslice.IsMatchesValue(h.excludeNetIPs, ifaceIPNet.String()) {
			continue
		}
		res = append(res, ifaceIP.String())
	}
	return res, nil
}
