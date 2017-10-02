package proxy

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/dh1tw/remoteRotator/rotator"
)

func TestInitial(t *testing.T) {

	host := Host("127.0.0.1")
	port := Port(7005)

	evh := func(r rotator.Rotator, ev rotator.Event, v ...interface{}) {
		log.Println(v)
	}

	evtHandler := EventHandler(evh)

	done := make(chan struct{})

	proxyRot, err := NewRotatorProxy(done, evtHandler, host, port)
	if err != nil {
		t.Fatal(err)
	}

	ticker := time.NewTicker(time.Second * 5)

	// Channel to handle OS signals
	osSignals := make(chan os.Signal, 1)

	//subscribe to os.Interrupt (CTRL-C signal)
	signal.Notify(osSignals, os.Interrupt)

	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-osSignals:
			t.Fatal("exiting...")
		case <-ticker.C:
			proxyRot.SetAzimuth(rand.Intn(360))
		case <-done:
			t.Fatal("done")
		}
	}

}
