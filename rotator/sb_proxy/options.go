package sbProxy

import (
	"github.com/asim/go-micro/v3/client"
	"github.com/dh1tw/remoteRotator/rotator"
)

func Client(cli client.Client) func(*SbProxy) {
	return func(r *SbProxy) {
		r.cli = cli
	}
}

// DoneCh is a functional option allows you to pass a channel to the proxy object.
// This channel will be closed by this object. It serves as a notification that
// the object can be deleted.
func DoneCh(ch chan struct{}) func(*SbProxy) {
	return func(r *SbProxy) {
		r.doneCh = ch
	}
}

func Name(name string) func(*SbProxy) {
	return func(r *SbProxy) {
		r.name = name
	}
}

func ServiceName(name string) func(*SbProxy) {
	return func(r *SbProxy) {
		r.serviceName = name
	}
}

// EventHandler sets a callback function through which the proxy rotator
// will report Events
func EventHandler(h func(rotator.Rotator, rotator.Heading)) func(*SbProxy) {
	return func(r *SbProxy) {
		r.eventHandler = h
	}
}
