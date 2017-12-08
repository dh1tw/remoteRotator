package discovery

import (
	"net"
	"strings"

	"github.com/micro/mdns"
)

// RotatorMdnsEntry is contains fields for a rotator discovered via mDNS.
type RotatorMdnsEntry struct {
	Name   string
	URL    string
	Host   string
	AddrV4 net.IP
	AddrV6 net.IP
	Port   int
}

// LookupRotators will perform an mDNS query are lookup all available
// rotators on the network.
// func LookupRotators() ([]RotatorMdnsEntry, error) {
// 	entriesCh := make(chan *mdns.ServiceEntry, 100)

func LookupRotators() ([]RotatorMdnsEntry, error) {
	entriesCh := make(chan *mdns.ServiceEntry, 100)

	rotators := []RotatorMdnsEntry{}

	go func() {
		for entry := range entriesCh {

			// ignore if not rotators.shackbus.local
			if !strings.Contains(entry.Name, "_rotator._tcp.local") {
				continue
			}

			name := strings.TrimSuffix(entry.Name, "._rotator._tcp.local.")
			// replace '\' (escaping backslashes)
			name = strings.Replace(name, "\x5c", "", -1)

			r := RotatorMdnsEntry{
				Name:   name,
				URL:    entry.Name,
				Host:   strings.TrimSuffix(entry.Host, "."),
				AddrV4: entry.AddrV4,
				AddrV6: entry.AddrV6,
				Port:   entry.Port,
			}
			rotators = append(rotators, r)
		}
	}()

	mdns.Lookup("_rotator._tcp", entriesCh)

	close(entriesCh)
	return rotators, nil
}
