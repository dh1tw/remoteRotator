package yaesu

import (
	"bytes"
	"testing"
	"time"
)

func TestSerialPortReadTimeout(t *testing.T) {

	// mock the serial port with Buffers
	dp := dummyPort{
		sendBuf: &bytes.Buffer{},
		rxBuf:   &bytes.Buffer{},
	}

	yaesu := Yaesu{
		sp:              &dp,
		closeCh:         make(chan struct{}),
		errorCh:         make(chan struct{}),
		pollingInterval: time.Millisecond * 100,
	}

	// after 5x pollingInterval the watchdog must kick in
	// to be on the safe side we wait for 7 pollingIntervals before failure
	timeout := time.After(yaesu.pollingInterval * 7)

	// the routine will try to query the rotator. But since we just
	// dump the content in buffer, there is no rotator replying. Hence
	// the Watchdog must kick in after 5x polling.
	go yaesu.start()
	select {
	case <-yaesu.errorCh:
		// expected behavior is that the watchdog will close
		// the error channel.
		yaesu.Close()
		timeout = nil
		time.Sleep(time.Second * 1)
		return
	case <-timeout:
		t.Fatal("Watchdog monitoring the serial port did not launch on read timeout")
	}

	// workaround - time for teardown needed - otherwise the test crashes on macos
	dp.Close()
}
