package proxy

import "github.com/dh1tw/remoteRotator/rotator"

// Host is a functional option to set IP / dns name of the remote Rotators host.
func Host(host string) func(*Proxy) {
	return func(r *Proxy) {
		r.host = host
	}
}

// Port is a functional option to set port of the remote Rotators on its host.
func Port(port int) func(*Proxy) {
	return func(r *Proxy) {
		r.port = port
	}
}

// DoneCh is a functional option allows you to pass a channel to the proxy object.
// The channel will be closed and thus notifies you when the object has been deleted.
func DoneCh(ch chan struct{}) func(*Proxy) {
	return func(r *Proxy) {
		r.doneCh = ch
	}
}

// EventHandler sets a callback function through which the proxy rotator
// will report Events
func EventHandler(h func(rotator.Rotator, rotator.Heading)) func(*Proxy) {
	return func(r *Proxy) {
		r.eventHandler = h
	}
}
