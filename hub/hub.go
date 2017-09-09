package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/GeertJohan/go.rice"

	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/gorilla/websocket"
)

// Hub is a struct which makes a rotator available through network
// interfaces, supporting several protocols.
type Hub struct {
	sync.Mutex
	tcpClients     map[*TCPClient]bool
	closeTCPClient chan *TCPClient
	wsClients      map[*WsClient]bool
	closeWsClient  chan *WsClient
	rotator        rotator.Rotator
}

// NewHub returns the pointer to an initialized Hub object for a
// given rotator.
func NewHub(r rotator.Rotator) *Hub {
	hub := &Hub{
		tcpClients:     make(map[*TCPClient]bool),
		closeTCPClient: make(chan *TCPClient),
		wsClients:      make(map[*WsClient]bool),
		closeWsClient:  make(chan *WsClient),
		rotator:        r,
	}

	go hub.handleClose()

	return hub
}

func (hub *Hub) handleClose() {
	for {
		select {
		case c := <-hub.closeTCPClient:
			hub.RemoveTCPClient(c)
		case c := <-hub.closeWsClient:
			hub.RemoveWsClient(c)
		}
	}
}

// AddTCPClient registers a new tcp client
func (hub *Hub) AddTCPClient(client *TCPClient) {
	hub.Lock()
	defer hub.Unlock()

	if _, alreadyInMap := hub.tcpClients[client]; alreadyInMap {
		delete(hub.tcpClients, client)
	}
	hub.tcpClients[client] = true
	// start listening on TCP socket
	log.Printf("tcp client connected (%v)\n", client.RemoteAddr())
	go client.listen(hub.rotator, hub.closeTCPClient)
}

// RemoveTCPClient removes a tcp client
func (hub *Hub) RemoveTCPClient(c *TCPClient) {
	hub.Lock()
	defer hub.Unlock()

	if _, ok := hub.tcpClients[c]; ok {
		delete(hub.tcpClients, c)
	}

	c.Close()
	log.Printf("tcp client disconnected (%v)\n", c.RemoteAddr())
}

// AddWsClient registers a new tcp client
func (hub *Hub) AddWsClient(client *WsClient) {

	if _, alreadyInMap := hub.wsClients[client]; alreadyInMap {
		delete(hub.wsClients, client)
	}
	hub.wsClients[client] = true
	// TBD: Start listening on websocket
	log.Printf("websocket client connected (%v)\n", client.RemoteAddr())
	go client.listen(hub.rotator, hub.closeWsClient)
}

// RemoveWsClient removes a tcp client
func (hub *Hub) RemoveWsClient(c *WsClient) {
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
		log.Printf("tcp listener error: %v", err.Error())
	}

	// Close the listener when the application closes.
	defer l.Close()

	fmt.Printf("Listening on %s:%d for TCP connections\n", host, port)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			// os.Exit(1)
		}

		c := &TCPClient{
			Conn: conn,
		}
		hub.AddTCPClient(c)
	}
}

func (hub *Hub) wsHandler(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := &WsClient{
		Conn: conn,
	}

	s := hub.rotator.Serialize()
	c.write(s)

	hub.AddWsClient(c)
}

func (hub *Hub) infoHandler(w http.ResponseWriter, r *http.Request) {
	i := []rotator.Info{hub.rotator.Info()}

	data, err := json.Marshal(i)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(data))
}

// ListenHTTP starts a HTTP Server on a given network adapter / port and
// sets a HTTP and Websocket handler.
// Since this function contains an endless loop, it should be executed
// in a go routine. If the listener can not be initialized, it will
// close the wsError channel.
func (hub *Hub) ListenHTTP(host string, port int, wsError chan<- bool) {

	defer close(wsError)

	box := rice.MustFindBox("../html")
	fileServer := http.FileServer(box.HTTPBox())

	http.Handle("/", fileServer)
	http.HandleFunc("/info", hub.infoHandler)
	http.HandleFunc("/ws", hub.wsHandler)

	// Listen for incoming connections.
	fmt.Printf("Listening on %s:%d for HTTP connections\n", host, port)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		log.Println(err)
		return
	}
}

// Broadcast sends a rotator Status struct to all connected clients
func (hub *Hub) Broadcast(s rotator.Status) {

	hub.BroadcastToTCPClients(s)
	if err := hub.BroadcastToWsClients(s); err != nil {
		log.Println(err)
	}
}

// BroadcastToTCPClients will send a rotator.Status struct to all connected
// TCP Clients
func (hub *Hub) BroadcastToTCPClients(s rotator.Status) {
	hub.Lock()
	defer hub.Unlock()

	// update the tcp Clients
	for c := range hub.tcpClients {
		// EA4TX's ARSVCOM doesn't understand single Azimuth
		// messages (+0nnn). It always expects +0nnn+0nnn
		data := fmt.Sprintf("+0%.3d+0%.3d\r\n", s.Azimuth, s.Elevation)
		if err := c.write(data); err != nil {
			log.Printf("error writing to client %v: %v\n", c.RemoteAddr(), err)
			log.Printf("disconnecting client %v\n", c.RemoteAddr())
			c.Close()
			delete(hub.tcpClients, c)
		}
	}
}

// BroadcastToWsClients will send a rotator.Status struct to all clients
// connected through a Websocket
func (hub *Hub) BroadcastToWsClients(s rotator.Status) error {
	hub.Lock()
	defer hub.Unlock()

	for c := range hub.wsClients {
		if err := c.write(s); err != nil {
			log.Printf("error writing to client %v: %v\n", c.RemoteAddr(), err)
			log.Printf("disconnecting client %v\n", c.RemoteAddr())
			c.Close()
			delete(hub.wsClients, c)
		}
	}

	return nil
}
