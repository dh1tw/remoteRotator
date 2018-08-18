package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/dh1tw/remoteRotator/hub"
	"github.com/dh1tw/remoteRotator/rotator"
)

const (
	// Time allowed to write a message to the peer.
	wsWriteWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	wsPongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	wsPingPeriod = 3 * time.Second
)

// Proxy is a proxy object representing a remote rotator. It implements
// the rotator.Rotator interface. Behind the scenes it sychronizes itself
// with the real rotator through a websocket.
type Proxy struct {
	sync.RWMutex
	host           string
	port           int
	conn           *websocket.Conn
	wsTxTimeout    time.Duration
	wsRxTimeout    time.Duration
	eventHandler   func(rotator.Rotator, rotator.Heading)
	name           string
	azimuthMin     int
	azimuthMax     int
	azimuthStop    int
	azimuthOverlap bool
	elevationMin   int
	elevationMax   int
	hasAzimuth     bool
	hasElevation   bool
	azimuth        int
	azPreset       int
	elevation      int
	elPreset       int
	closeCh        chan struct{}
	doneCh         chan struct{}
}

// New returns the pointer to an initalized Rotator proxy object.
func New(opts ...func(*Proxy)) (*Proxy, error) {

	r := &Proxy{
		name:    "rotatorProxy",
		closeCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(r)
	}

	if err := r.getObject(); err != nil {
		return nil, err
	}

	wsDialer := &websocket.Dialer{}

	wsURL := fmt.Sprintf("ws://%s:%d/ws", r.host, r.port)
	conn, _, err := wsDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	// Pong handler extends the read deadline by wsPongWait whenever a
	// pong has been received
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	r.conn = conn

	// this function sends every wsPingPeriod a ping to the other side.
	// if this fails, the function terminates. No further signaling needed,
	// since the readTimeout will kick in eventually and start the object
	// disposal.
	go func() {
		ping := time.NewTicker(wsPingPeriod)
		for {
			<-ping.C
			r.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := r.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}()

	// this function listen on the websocket for incoming messages or until
	// readTimeout kicks in. This shouldn't happen as long as the counterpart
	// responds to the pings.
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					log.Println("websocket error:", err)
				}
				// Signal the object holder that we are going to shutdown so
				// that this object can be disposed.
				close(r.doneCh)
				return
			}

			data := hub.Event{}
			if err := json.Unmarshal(msg, &data); err != nil {
				log.Println(err)
			}

			switch data.Name {
			case "add":
				// pass
			case "remove":
				// pass
			case "heading":
				r.Lock()
				changed := false

				s := data.Heading
				if r.azimuth != s.Azimuth {
					r.azimuth = s.Azimuth
					changed = true
				}
				if r.azPreset != s.AzPreset {
					r.azPreset = s.AzPreset
					changed = true
				}
				if r.elevation != s.Elevation {
					r.elevation = s.Elevation
					changed = true
				}
				if r.elPreset != s.ElPreset {
					r.elPreset = s.ElPreset
					changed = true
				}

				if changed {
					if r.eventHandler != nil {
						go r.eventHandler(r, s)
					}
				}
				r.Unlock()
			}
		}
	}()

	return r, nil
}

func (r *Proxy) Close() {
	// if r.conn != nil {
	// 	r.conn.Close()
	// }
}

// get the serialized representation of the local rotator object and set the
// same parameters in our proxy Object
func (r *Proxy) getObject() error {

	url := fmt.Sprintf("http://%s:%d/api/rotators", r.host, r.port)

	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rotators := rotator.Objects{}

	if err := json.NewDecoder(resp.Body).Decode(&rotators); err != nil {
		return err
	}

	if len(rotators) == 0 {
		return fmt.Errorf("incompatible rotator at %v:%v", r.host, r.port)
	}

	if len(rotators) > 1 {
		return fmt.Errorf("expected information of 1 rotator, but got %d", len(rotators))
	}

	// there is only one rotator in the dict
	for _, pr := range rotators {
		r.name = pr.Name
		r.hasAzimuth = pr.Config.HasAzimuth
		r.hasElevation = pr.Config.HasElevation
		r.azimuthMin = pr.Config.AzimuthMin
		r.azimuthMax = pr.Config.AzimuthMax
		r.azimuthStop = pr.Config.AzimuthStop
		r.elevationMin = pr.Config.ElevationMin
		r.elevationMax = pr.Config.ElevationMax
		r.azimuth = pr.Heading.Azimuth
		r.azPreset = pr.Heading.AzPreset
		r.elevation = pr.Heading.Elevation
		r.elPreset = pr.Heading.ElPreset
	}

	return nil
}

func (r *Proxy) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.name
}

func (r *Proxy) HasAzimuth() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasAzimuth
}

func (r *Proxy) HasElevation() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasElevation
}

func (r *Proxy) Azimuth() int {
	r.RLock()
	defer r.RUnlock()
	return r.azimuth
}

func (r *Proxy) AzPreset() int {
	r.RLock()
	defer r.RUnlock()
	return r.azPreset
}

func (r *Proxy) SetAzimuth(az int) error {

	azPut := rotator.AzimuthPut{
		Azimuth: &az,
	}

	url := fmt.Sprintf("http://%s:%d/api/rotator/%s/azimuth", r.host, r.port, r.name)

	return putRequest(url, &azPut)
}

func (r *Proxy) Elevation() int {
	r.RLock()
	defer r.RUnlock()
	return r.elevation
}

func (r *Proxy) ElPreset() int {
	r.RLock()
	defer r.RUnlock()
	return r.elPreset
}

func (r *Proxy) SetElevation(el int) error {

	elPut := rotator.ElevationPut{
		Elevation: &el,
	}

	url := fmt.Sprintf("http://%s:%d/api/rotator/%s/elevation", r.host, r.port, r.name)

	return putRequest(url, &elPut)
}

func (r *Proxy) StopAzimuth() error {

	url := fmt.Sprintf("http://%s:%d/api/rotator/%s/stop_azimuth", r.host, r.port, r.name)

	return putRequest(url, struct{}{})
}

func (r *Proxy) StopElevation() error {
	url := fmt.Sprintf("http://%s:%d/api/rotator/%s/stop_elevation", r.host, r.port, r.name)

	return putRequest(url, struct{}{})
}

func (r *Proxy) Stop() error {
	url := fmt.Sprintf("http://%s:%d/api/rotator/%s/stop", r.host, r.port, r.name)

	return putRequest(url, struct{}{})
}

// Serialize the data of the rotator
func (r *Proxy) Serialize() rotator.Object {
	r.RLock()
	defer r.RUnlock()
	return r.serialize()
}

func (r *Proxy) serialize() rotator.Object {

	obj := rotator.Object{
		Name: r.name,
		Heading: rotator.Heading{
			Azimuth:   int(r.azimuth),
			AzPreset:  int(r.azPreset),
			Elevation: int(r.elevation),
			ElPreset:  int(r.elPreset),
		},
		Config: rotator.Config{
			HasAzimuth:   r.hasAzimuth,
			HasElevation: r.hasElevation,
			AzimuthMax:   r.azimuthMax,
			AzimuthMin:   r.azimuthMin,
			AzimuthStop:  r.azimuthStop,
			ElevationMax: r.elevationMax,
			ElevationMin: r.elevationMin,
		},
	}

	return obj
}

// putRequest executes an HTTP put request.
func putRequest(url string, data interface{}) error {

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(data)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := http.NewRequest("PUT", url, b)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return (err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to set azimuth. http error code is %v", resp.StatusCode)
	}

	return nil
}
