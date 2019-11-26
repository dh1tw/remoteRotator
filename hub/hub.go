package hub

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"sync"

	rice "github.com/GeertJohan/go.rice"

	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/gorilla/mux"
)

// Hub is a struct which makes a rotator available through network
// interfaces, supporting several protocols.
type Hub struct {
	sync.RWMutex
	tcpClients     map[*TCPClient]bool
	closeTCPClient chan *TCPClient
	wsClients      map[*WsClient]bool
	closeWsClient  chan *WsClient
	rotators       map[string]rotator.Rotator //key: Rotator name
	router         *mux.Router
	fileServer     http.Handler
	apiVersion     string
	apiMatch       *regexp.Regexp
}

// NewHub returns the pointer to an initialized Hub object.
func NewHub(rotators ...rotator.Rotator) (*Hub, error) {
	hub := &Hub{
		tcpClients:     make(map[*TCPClient]bool),
		closeTCPClient: make(chan *TCPClient),
		wsClients:      make(map[*WsClient]bool),
		closeWsClient:  make(chan *WsClient),
		rotators:       make(map[string]rotator.Rotator),
		apiVersion:     "1.0",
		apiMatch:       regexp.MustCompile(`api\/v\d\.\d\/`),
	}

	for _, r := range rotators {
		if err := hub.AddRotator(r); err != nil {
			return nil, err
		}
	}

	go hub.handleClose()

	return hub, nil
}

func (hub *Hub) handleClose() {
	for {
		select {
		case c := <-hub.closeTCPClient:
			hub.removeTCPClient(c)
		case c := <-hub.closeWsClient:
			hub.removeWsClient(c)
		}
	}
}

// AddRotator adds / registers a rotator. The rotator's name must be unique.
func (hub *Hub) AddRotator(r rotator.Rotator) error {
	hub.Lock()
	defer hub.Unlock()

	return hub.addRotator(r)
}

func (hub *Hub) addRotator(r rotator.Rotator) error {
	_, ok := hub.rotators[r.Name()]
	if ok {
		return fmt.Errorf("rotator names must be unique; %s exists more than once", r.Name())
	}
	hub.rotators[r.Name()] = r
	ev := Event{
		Name:        AddRotator,
		RotatorName: r.Name(),
	}
	hub.broadcast(ev)
	log.Printf("added rotator (%s)\n", r.Name())

	return nil
}

// RemoveRotator deletes / de-registers a rotator.
func (hub *Hub) RemoveRotator(r rotator.Rotator) {
	hub.Lock()
	defer hub.Unlock()

	ev := Event{
		Name:        RemoveRotator,
		RotatorName: r.Name(),
	}

	hub.broadcast(ev)

	r.Close()
	delete(hub.rotators, r.Name())
	log.Printf("removed rotator (%s)\n", r.Name())
}

// Rotator returns a particular rotator stored from the hub. If no
// rotator exists with that name, (nil, false) will be returned.
func (hub *Hub) Rotator(name string) (rotator.Rotator, bool) {
	hub.RLock()
	defer hub.RUnlock()

	rotator, ok := hub.rotators[name]
	return rotator, ok
}

// Rotators returns a slice of all registered rotators.
func (hub *Hub) Rotators() []rotator.Rotator {
	hub.RLock()
	defer hub.RUnlock()

	rotators := make([]rotator.Rotator, 0, len(hub.rotators))
	for _, r := range hub.rotators {
		rotators = append(rotators, r)
	}

	return rotators
}

// addTCPClient registers a new tcp client
func (hub *Hub) addTCPClient(client *TCPClient) {
	hub.Lock()
	defer hub.Unlock()

	if _, alreadyInMap := hub.tcpClients[client]; alreadyInMap {
		delete(hub.tcpClients, client)
	}
	hub.tcpClients[client] = true
	// start listening on TCP socket
	log.Printf("tcp client connected (%v)\n", client.RemoteAddr())

	// we always pick the first rotator since the TCP client implements
	// the Yaesu GS232 protocol which can only talk to a single rotator.
	for _, r := range hub.rotators {
		go client.listen(r, hub.closeTCPClient)
		break
	}
}

// RemoveTCPClient removes a tcp client
func (hub *Hub) removeTCPClient(c *TCPClient) {
	hub.Lock()
	defer hub.Unlock()

	if _, ok := hub.tcpClients[c]; ok {
		delete(hub.tcpClients, c)
	}

	c.Close()
	log.Printf("tcp client disconnected (%v)\n", c.RemoteAddr())
}

// AddWsClient registers a new websocket client
func (hub *Hub) addWsClient(client *WsClient) {
	hub.Lock()
	defer hub.Unlock()

	if _, alreadyInMap := hub.wsClients[client]; alreadyInMap {
		delete(hub.wsClients, client)
	}
	hub.wsClients[client] = true

	// we need to listen on the websocket so that the incoming ping
	// messages can be (automatically) answered (with a pong message)
	go client.listen(hub.closeWsClient)

	log.Printf("websocket client connected (%v)\n", client.RemoteAddr())
}

// removeWsClient removes a websocket client
func (hub *Hub) removeWsClient(c *WsClient) {
	hub.Lock()
	defer hub.Unlock()

	if _, ok := hub.wsClients[c]; ok {
		delete(hub.wsClients, c)
	}

	c.Close()
	log.Printf("websocket client disconnected (%v)\n", c.RemoteAddr())
}

// ListenTCP starts a TCP listener on a given network adapter / port.
// Since this function contains an endless loop, it should be executed
// in a go routine. If the listener can not be initialized, it will
// close the tcpError channel.
func (hub *Hub) ListenTCP(host string, port int, tcpError chan<- bool) {
	defer close(tcpError)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Printf("tcp listener error (%v)", err.Error())
		return
	}

	// Close the listener when the application closes.
	defer l.Close()

	log.Printf("listening on %s:%d for TCP connections\n", host, port)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println("error accepting: ", err.Error())
		}

		c := &TCPClient{
			Conn: conn,
		}
		hub.addTCPClient(c)
	}
}

// ListenHTTP starts a HTTP Server on a given network adapter / port and
// sets a HTTP and Websocket handler.
// Since this function contains an endless loop, it should be executed
// in a go routine. If the listener can not be initialized, it will
// close the errorCh channel.
func (hub *Hub) ListenHTTP(host string, port int, errorCh chan<- struct{}) {

	defer close(errorCh)

	box := rice.MustFindBox("../html")
	hub.fileServer = http.FileServer(box.HTTPBox())
	hub.router = mux.NewRouter().StrictSlash(true)

	// load the HTTP routes with their respective endpoints
	hub.routes()

	// Listen for incoming connections.
	log.Printf("listening on %s:%d for HTTP connections\n", host, port)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), hub.apiRedirectRouter(hub.router))
	if err != nil {
		log.Println(err)
		return
	}
}

// Broadcast sends a rotator event to all connected clients
func (hub *Hub) Broadcast(ev Event) {
	hub.Lock()
	defer hub.Unlock()
	hub.broadcast(ev)
}

func (hub *Hub) broadcast(ev Event) {
	hub.broadcastToTCPClients(ev)
	hub.broadcastToWsClients(ev)
}

// BroadcastToTCPClients will send a rotator event to all connected TCP Clients
func (hub *Hub) broadcastToTCPClients(ev Event) {

	// TCP only supports Heading messages
	if ev.Name != UpdateHeading {
		return
	}

	// update the tcp Clients
	for c := range hub.tcpClients {
		// EA4TX's ARSVCOM doesn't understand single Azimuth
		// messages (+0nnn). It always expects +0nnn+0nnn
		data := fmt.Sprintf("+0%.3d+0%.3d\r\n", ev.Heading.Azimuth, ev.Heading.Elevation)
		if err := c.write(data); err != nil {
			log.Printf("error writing to client %v: %v\n", c.RemoteAddr(), err)
			log.Printf("disconnecting client %v\n", c.RemoteAddr())
			c.Close()
			delete(hub.tcpClients, c)
		}
	}
}

type Event struct {
	Name        RotatorEvent    `json:"name,omitempty"`
	RotatorName string          `json:"rotator_name,omitempty"`
	Heading     rotator.Heading `json:"heading,omitempty"`
}

type RotatorEvent string

const (
	AddRotator    RotatorEvent = "add"
	RemoveRotator RotatorEvent = "remove"
	UpdateHeading RotatorEvent = "heading"
)

func (hub *Hub) broadcastToWsClients(event Event) {

	for c := range hub.wsClients {
		if err := c.write(event); err != nil {
			log.Printf("error writing to client %v: %v\n", c.RemoteAddr(), err)
			log.Printf("disconnecting client %v\n", c.RemoteAddr())
			c.Close()
			delete(hub.wsClients, c)
		}
	}
}
