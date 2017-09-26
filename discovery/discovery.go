package discovery

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/micro/mdns"
)

// RotatorMdnsEntry is contains fields for a rotator discovered via mDNS.
type RotatorMdnsEntry struct {
	rotator.Info
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
			r := RotatorMdnsEntry{
				URL:    entry.Name,
				Host:   entry.Host,
				AddrV4: entry.AddrV4,
				AddrV6: entry.AddrV6,
				Port:   entry.Port,
			}
			if len(entry.InfoFields) > 0 {
				info, err := decodeInfo(entry.InfoFields[0])
				if err != nil {
					fmt.Printf("invalid txt record of %v: %v\n", entry.Name, err)
					r.Name = "unknown"
				} else {
					r.Info = info
				}
			} else {
				fmt.Printf("no txt record found for %v\n", entry.Name)
				r.Name = "unknown"
			}
			rotators = append(rotators, r)
		}
	}()

	// Start the lookup
	mdns.Lookup("rotators.shackbus", entriesCh)

	close(entriesCh)
	return rotators
}

// Rotators return a rotator.Info struct, embedded as a TXT record in
// their mDNS response. The rotator.Info struct is serialized (JSON) and encoded
// with base64. This function decodes the message and returns the deserialized
// rotator.Info struct.
func decodeInfo(rawB64 string) (rotator.Info, error) {

	info := rotator.Info{}

	uDec, err := b64.URLEncoding.DecodeString(rawB64)
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(uDec, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}
