package main

import (
	"errors"
	"fmt"
	"net"
)

func main() {

	ip, err := findSystemIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ip)

}

func findSystemIP() (string, error) {
	// list of system network interfaces
	// https://golang.org/pkg/net/#Interfaces
	intfs, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	// mapping between network interface name and index
	// https://golang.org/pkg/net/#Interface
	for _, intf := range intfs {
		fmt.Println(intf.Name + ":")
		// skip down interface & check next intf
		if intf.Flags&net.FlagUp == 0 {
			continue
		}
		// skip loopback & check next intf
		if intf.Flags&net.FlagLoopback != 0 {
			continue
		}
		// list of unicast interface addresses for specific interface
		// https://golang.org/pkg/net/#Interface.Addrs
		addrs, err := intf.Addrs()
		if err != nil {
			return "", err
		}
		// network end point address
		// https://golang.org/pkg/net/#Addr
		for _, addr := range addrs {
			var ip net.IP
			// Addr type switch required as a result of IPNet & IPAddr return in
			// https://golang.org/src/net/interface_windows.go?h=interfaceAddrTable
			switch v := addr.(type) {
			// net.IPNet satisfies Addr interface
			// since it contains Network() & String()
			// https://golang.org/pkg/net/#IPNet
			case *net.IPNet:
				ip = v.IP
			// net.IPAddr satisfies Addr interface
			// since it contains Network() & String()
			// https://golang.org/pkg/net/#IPAddr
			case *net.IPAddr:
				ip = v.IP
			}
			// skip loopback & check next addr
			if ip == nil || ip.IsLoopback() {
				continue
			}
			// convert IP IPv4 address to 4-byte
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			// return IP address as string
			return ip.String(), nil
		}
	}
	return "", errors.New("no ip interface up")
}
