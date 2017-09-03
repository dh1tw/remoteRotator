package hub

import (
	"fmt"
	"log"
	"net"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/dh1tw/remoteRotator/rotator"
)

//WsClient is a wrapper for clients connected through a Websocket
type WsClient struct {
	Ws *websocket.Conn
}

// Hub is a struct which makes a rotator available through network
// interfaces, supporting several protocols.
type Hub struct {
	sync.Mutex
	tcpClients map[*TCPClient]bool
	wsClients  map[*WsClient]bool
	rotator    rotator.Rotator
}

// NewHub returns the pointer to an initialized Hub object for a
// given rotator.
func NewHub(r rotator.Rotator) *Hub {
	hub := &Hub{
		tcpClients: make(map[*TCPClient]bool),
		wsClients:  make(map[*WsClient]bool),
		rotator:    r,
	}

	return hub
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
	log.Printf("client connected (%v)\n", client.Conn.RemoteAddr())
	go client.listen(hub)
}

// AddWsClient registers a new tcp client
func (hub *Hub) AddWsClient(client *WsClient) {

	if _, alreadyInMap := hub.wsClients[client]; alreadyInMap {
		delete(hub.wsClients, client)
	}
	hub.wsClients[client] = true
	// TBD: Start listening on websocket
	log.Printf("websocket client connected (%v)\n", client.Ws.RemoteAddr())
}

// RemoveTCPClient removes a tcp client
func (hub *Hub) RemoveTCPClient(c *TCPClient) {
	hub.Lock()
	defer hub.Unlock()

	if _, ok := hub.tcpClients[c]; ok {
		delete(hub.tcpClients, c)
		c.Conn.Close()
		log.Printf("client disconnected (%v)\n", c.Conn.RemoteAddr())
	}
}

// ListenTCP starts a TCP listener on a given network adapter / port.
// Since this function contains an endless loop, it should be executed
// in a go routine. If the listener can not be initialized, it will
// close the tcpError channel.
func (hub *Hub) ListenTCP(host string, port int, tcpError chan<- bool) error {
	// Listen for incoming connections.
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("tcp listener error: %v", err.Error())
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

		fmt.Println(conn.RemoteAddr())

		c := &TCPClient{
			Conn: conn,
		}
		hub.AddTCPClient(c)
	}
}

// Broadcast sends data to all connected clients
func (hub *Hub) Broadcast(s rotator.Status) {
	hub.Lock()
	defer hub.Unlock()

	// update the tcp Clients
	for c := range hub.tcpClients {
		// EA4TX's ARSVCOM doesn't understand single Azimuth
		// messages (+0nnn). It always expects +0nnn+0nnn
		data := fmt.Sprintf("+0%.3d+0%.3d\r\n", s.Azimuth, s.Elevation)
		if err := c.write(data); err != nil {
			log.Printf("error writing to client %v: %v\n", c.Conn.RemoteAddr(), err)
			log.Printf("disconnecting client %v\n", c.Conn.RemoteAddr())
			c.Conn.Close()
			delete(hub.tcpClients, c)
		}
	}

	for _ = range hub.wsClients {
		// TBD ... same for websockets
	}
}
