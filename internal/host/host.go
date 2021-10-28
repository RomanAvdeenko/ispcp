package host

import (
	"log"
	"net"
	"strings"
)

// Host implements unix host...
type Host struct {
	processedInterfaces *[]net.Interface
	excludeIfaceNames   *[]string
}

func NewHost() *Host {
	return &Host{}
}

func (h *Host) SetExcludeInterfaceNames(val []string) {
	if val != nil {
		h.excludeIfaceNames = &val
	}

}

func (h *Host) Configure() {
	// get Interfaces exclude unprocessed
	ifaces, _ := net.Interfaces()
	// Filter
	filteredIfaces := []net.Interface{}
	for _, iface := range ifaces {
		log.Printf("%v: %v\n", iface.Name, iface.Flags)
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
		for _, excludeVal := range *h.excludeIfaceNames {
			if strings.HasPrefix(iface.Name, excludeVal) {
				needContinue = true
				break
			}
		}
		if needContinue {
			continue
		}
		//
		filteredIfaces = append(filteredIfaces, iface)
	}
	h.processedInterfaces = &filteredIfaces
	log.Printf("h.processedInterfaces: %#v\n", h.processedInterfaces)
}

//func (h * Host)
