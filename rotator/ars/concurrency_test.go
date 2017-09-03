package ars

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// randPort mocks the serial port for this test.
// randPort implements the ReadWriteCloser Interface
type randPort struct {
	bytesSent uint64    // atomic counter for bytes sent to ARS via serial port
	bytesRcvd uint64    // atomic counter for bytes recevied from ARS via serial port
	ts        time.Time // timestamp to limit calls
}

// simulate traffic sent from ARS to our application
// every 50ms
func (p *randPort) Read(b []byte) (int, error) {
	if time.Since(p.ts) > time.Millisecond*50 {
		p.ts = time.Now()
		az := rand.Intn(450)
		el := rand.Intn(180)
		b = []byte(fmt.Sprintf("+%3d+%d", az, el))
		atomic.AddUint64(&p.bytesRcvd, uint64(len(b)))
		return len(b), nil
	}
	return 0, io.EOF //simulate Timeout error
}

func (p *randPort) Write(b []byte) (int, error) {
	atomic.AddUint64(&p.bytesSent, uint64(len(b)))
	return len(b), nil
}

func (p *randPort) Close() error {
	return nil
}

type apiCallCounter struct {
	serialize     uint64
	azimuth       uint64
	elevation     uint64
	azPreset      uint64
	hasAzimuth    uint64
	hasElevation  uint64
	elPreset      uint64
	setAzimuth    uint64
	setElevation  uint64
	stop          uint64
	stopElevation uint64
	stopAzimuth   uint64
}

// This test will spin up 1000 go routines and call randomly all available
// Methods of the Ars object. The intention of this test is to detect any
// race conditions which could happen due to the concurrent access.
// A summary of the API calls and transfered bytes (rx/tx) after a successful
// pass.
func TestArsMassiveConcurrentCalls(t *testing.T) {
	dp := &randPort{}

	ars := &Ars{
		hasAzimuth:      true,
		sp:              dp,
		pollingInterval: time.Second * 2,
	}

	rand.Seed(time.Now().UTC().UnixNano())

	d := time.Second * 5
	wg := &sync.WaitGroup{}

	arsError := make(chan bool)
	shutdown := make(chan bool)

	calls := &apiCallCounter{}

	go ars.Start(arsError, shutdown)
	for i := 0; i < 1000; i++ {
		go randomAccess(ars, d, calls, wg, t)
		wg.Add(1)
	}

	select {
	case <-arsError:
		t.Errorf("unexpected error while reading from serial port")
	default:
	}
	wg.Wait()
	shutdown <- true
	time.Sleep(time.Second * 3)
	fmt.Println("Concurrent stress test summary:")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Printf("bytes sent to (fake ARS):       %d\n", atomic.LoadUint64(&dp.bytesSent))
	fmt.Printf("bytes received from (fake ARS): %d\n", atomic.LoadUint64(&dp.bytesRcvd))
	fmt.Printf("ars.Serialize called:     %d times\n", atomic.LoadUint64(&calls.serialize))
	fmt.Printf("ars.Azimuth called:       %d times\n", atomic.LoadUint64(&calls.azimuth))
	fmt.Printf("ars.Elevation called:     %d times\n", atomic.LoadUint64(&calls.elevation))
	fmt.Printf("ars.AzPreset called:      %d times\n", atomic.LoadUint64(&calls.azPreset))
	fmt.Printf("ars.HasAzimuth called:    %d times\n", atomic.LoadUint64(&calls.hasAzimuth))
	fmt.Printf("ars.HasElevation called:  %d times\n", atomic.LoadUint64(&calls.hasElevation))
	fmt.Printf("ars.ElPreset called:      %d times\n", atomic.LoadUint64(&calls.elPreset))
	fmt.Printf("ars.SetAzimuth called:    %d times\n", atomic.LoadUint64(&calls.setAzimuth))
	fmt.Printf("ars.SetElevation called:  %d times\n", atomic.LoadUint64(&calls.setElevation))
	fmt.Printf("ars.Stop called:          %d times\n", atomic.LoadUint64(&calls.stop))
	fmt.Printf("ars.StopAzimuth called:   %d times\n", atomic.LoadUint64(&calls.stopAzimuth))
	fmt.Printf("ars.StopElevation called: %d times\n", atomic.LoadUint64(&calls.stopElevation))
	// fmt.Println("write buffer:", dp.sendBuf.Len())
}

// just randomly call any of the API methods
func randomAccess(r *Ars, timeout time.Duration, c *apiCallCounter,
	wg *sync.WaitGroup, t *testing.T) {
	defer wg.Done()

	timeoutTimer := time.NewTimer(timeout)

	for {
		randFunc := rand.Intn(12)

		switch randFunc {
		case 0:
			r.Serialize()
			atomic.AddUint64(&c.serialize, 1)
		case 1:
			r.Azimuth()
			atomic.AddUint64(&c.azimuth, 1)
		case 2:
			r.Elevation()
			atomic.AddUint64(&c.elevation, 1)
		case 3:
			r.AzPreset()
			atomic.AddUint64(&c.azPreset, 1)
		case 4:
			r.HasAzimuth()
			atomic.AddUint64(&c.hasAzimuth, 1)
		case 5:
			r.HasElevation()
			atomic.AddUint64(&c.hasElevation, 1)
		case 6:
			r.ElPreset()
			atomic.AddUint64(&c.elPreset, 1)
		case 7:
			err := r.SetAzimuth(rand.Intn(450))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			atomic.AddUint64(&c.setAzimuth, 1)
		case 8:
			err := r.SetElevation(rand.Intn(180))
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			atomic.AddUint64(&c.setElevation, 1)
		case 9:
			err := r.Stop()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			atomic.AddUint64(&c.stop, 1)
		case 10:
			err := r.StopElevation()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			atomic.AddUint64(&c.stopElevation, 1)
		case 11:
			err := r.StopAzimuth()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			atomic.AddUint64(&c.stopAzimuth, 1)
		}

		select {
		case <-timeoutTimer.C:
			return
		default:
			//pass
		}
	}
}

// input range 0...500
func randBool(v int) bool {
	if v > 250 {
		return true
	}
	return false
}
