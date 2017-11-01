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
func LookupRotators() []RotatorMdnsEntry {
	entriesCh := make(chan *mdns.ServiceEntry, 100)

	// rotators := []*mdns.ServiceEntry{}
	rotators := []RotatorMdnsEntry{}

	go func() {
		for entry := range entriesCh {

			// ignore if not rotators.shackbus.local
			if !strings.Contains(entry.Name, "rotators.shackbus.local") {
				continue
			}

			name := strings.TrimSuffix(entry.Name, ".rotators.shackbus.local.")
			// replace '\' (escaping backslashes)
			name = strings.Replace(name, "\x5c", "", -1)

			r := RotatorMdnsEntry{
				Name:   name,
				URL:    entry.Name,
				Host:   entry.Host,
				AddrV4: entry.AddrV4,
				AddrV6: entry.AddrV6,
				Port:   entry.Port,
			}
			rotators = append(rotators, r)
		}
	}()

	// Start the lookup
	mdns.Lookup("rotators.shackbus", entriesCh)

	close(entriesCh)
	return rotators
}
