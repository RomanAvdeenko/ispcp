package host

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Host implements unix host...
type Host struct {
	processedInterfaces []net.Interface
	excludeIfaceNames   []string
	logger              *zerolog.Logger
}

func NewHost() *Host {
	return &Host{}
}

func (h *Host) SetExcludeInterfaceNames(val []string) {
	if val != nil {
		h.excludeIfaceNames = val
	}
}

func (h *Host) SetLogger(l *zerolog.Logger) { h.logger = l }

func (h *Host) Configure() {
	// Set default logger if needed
	if h.logger == nil {
		h.logger = &zerolog.Logger{}
		*h.logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
	}
	// Walk to interfaces
	ifaces, _ := net.Interfaces()
	processedInterfaces := []net.Interface{}
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
		var needContinue bool
		for _, excludeVal := range h.excludeIfaceNames {
			if strings.HasPrefix(iface.Name, excludeVal) {
				needContinue = true
				break
			}
		}
		if needContinue {
			continue
		}
		//
		processedInterfaces = append(processedInterfaces, iface)
	}
	h.processedInterfaces = processedInterfaces
	//
	h.logger.Debug().Msg("ProcessedInterfaces:")
	for _, val := range h.processedInterfaces {
		h.logger.Debug().Msg(val.Name)
	}

}
