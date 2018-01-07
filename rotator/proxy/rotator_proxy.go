package proxy

import (
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
	wsWriteMutex   sync.Mutex
	wsTxTimeout    time.Duration
	wsRxTimeout    time.Duration
	eventHandler   func(rotator.Rotator, rotator.Status)
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
func EventHandler(h func(rotator.Rotator, rotator.Status)) func(*Proxy) {
	return func(r *Proxy) {
		r.eventHandler = h
	}
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

	if err := r.getInfo(); err != nil {
		return nil, err
	}

	wsDialer := &websocket.Dialer{}

	wsURL := fmt.Sprintf("ws://%s:%d/ws", r.host, r.port)
	conn, _, err := wsDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	r.conn = conn

	go func() {
		ping := time.NewTicker(wsPingPeriod)
		for {
			select {
			case <-ping.C:
				r.wsWriteMutex.Lock()
				r.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := r.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					// log.Println(err)
					r.wsWriteMutex.Unlock()
					return
				}
				r.wsWriteMutex.Unlock()
			}
		}
	}()

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					log.Println("websocket error:", err)
				}
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
				s := data.Status
				if r.azimuth != s.Azimuth {
					r.azimuth = s.Azimuth
					if r.eventHandler != nil {
						go r.eventHandler(r, s)
					}
				}
				if r.azPreset != s.AzPreset {
					r.azPreset = s.AzPreset
					if r.eventHandler != nil {
						go r.eventHandler(r, s)
					}
				}
				if r.elevation != s.Elevation {
					r.elevation = s.Elevation
					if r.eventHandler != nil {
						go r.eventHandler(r, s)
					}
				}
				if r.elPreset != s.ElPreset {
					r.elPreset = s.ElPreset
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
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *Proxy) getInfo() error {
	infoURL := fmt.Sprintf("http://%s:%d/info", r.host, r.port)

	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(infoURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	infos := []rotator.Info{}

	if err := json.NewDecoder(resp.Body).Decode(&infos); err != nil {
		return err
	}

	if len(infos) > 1 {
		return fmt.Errorf("expected information of 1 rotator, but got %d", len(infos))
	}

	r.name = infos[0].Name
	r.hasAzimuth = infos[0].HasAzimuth
	r.hasElevation = infos[0].HasElevation
	r.azimuthMin = infos[0].AzimuthMin
	r.azimuthMax = infos[0].AzimuthMax
	r.azimuthStop = infos[0].AzimuthStop
	r.elevationMin = infos[0].ElevationMin
	r.elevationMax = infos[0].ElevationMax
	r.azimuth = infos[0].Azimuth
	r.azPreset = infos[0].AzPreset
	r.elevation = infos[0].Elevation
	r.elPreset = infos[0].ElPreset

	return nil
}

func (r *Proxy) write(s rotator.Status) error {
	r.wsWriteMutex.Lock()
	defer r.wsWriteMutex.Unlock()
	return r.conn.WriteJSON(s)
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
	req := rotator.Request{
		HasAzimuth: true,
		Azimuth:    az,
	}

	return r.conn.WriteJSON(req)
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
	req := rotator.Request{
		HasElevation: true,
		Elevation:    el,
	}

	return r.conn.WriteJSON(req)
}

func (r *Proxy) StopAzimuth() error {
	req := rotator.Request{
		StopAzimuth: true,
	}

	return r.conn.WriteJSON(req)
}

func (r *Proxy) StopElevation() error {
	req := rotator.Request{
		StopElevation: true,
	}

	return r.conn.WriteJSON(req)
}

func (r *Proxy) Stop() error {
	req := rotator.Request{
		Stop: true,
	}

	return r.conn.WriteJSON(req)
}

func (r *Proxy) Status() rotator.Status {
	r.RLock()
	defer r.RUnlock()

	return rotator.Status{
		Name:           r.name,
		Azimuth:        r.azimuth,
		AzPreset:       r.azPreset,
		AzimuthOverlap: r.azimuthOverlap,
		Elevation:      r.elevation,
		ElPreset:       r.elPreset,
	}
}

func (r *Proxy) ExecuteRequest(req rotator.Request) error {
	return r.conn.WriteJSON(req)
}

func (r *Proxy) Info() rotator.Info {
	r.RLock()
	defer r.RUnlock()

	return rotator.Info{
		Name:           r.name,
		HasAzimuth:     r.hasAzimuth,
		HasElevation:   r.hasElevation,
		AzimuthMin:     r.azimuthMin,
		AzimuthMax:     r.azimuthMax,
		AzimuthStop:    r.azimuthStop,
		AzimuthOverlap: r.azimuthOverlap,
		ElevationMin:   r.elevationMin,
		ElevationMax:   r.elevationMax,
		Azimuth:        r.azimuth,
		AzPreset:       r.azPreset,
		Elevation:      r.elevation,
		ElPreset:       r.elPreset,
	}
}
