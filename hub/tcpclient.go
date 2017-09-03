package hub

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

//TCPClient is a wrapper for clients connected through plain a TCP socket.
type TCPClient struct {
	Conn net.Conn
}

// listen starts listening for incoming messages from tcp connections. This
// function is typically executed in a go routine to avoid blocking. When
// a error occurs, the routine returns and deletes the tcp connection.
// Since this method contains an endless loop it should be executed
// in a go routine.
func (c *TCPClient) listen(hub *Hub) {
	defer hub.RemoveTCPClient(c)

	for {
		msg, err := bufio.NewReader(c.Conn).ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("socket read error (%v): %v\n", c.Conn.RemoteAddr(), err)
			}
			return //disconnect and remove client
		}

		switch strings.ToUpper(msg[0:1]) {
		// set azimuth / elevation heading
		case "M":
			msg = strings.TrimRight(msg[1:], "\r\n")

			if len(msg) == 0 {
				if err := c.prompt(); err != nil {
					log.Println(err)
					return
				}
				continue
			}
			// TBD need to handle elevation
			az, err := strconv.Atoi(msg)
			if err != nil {
				log.Printf("parse error (%v): %v; msg: %s\n", c.Conn.RemoteAddr(), err, msg)
				continue
			}
			hub.rotator.SetAzimuth(az)
		// query
		case "C":
			// azimuth + elevation
			if msg[1] == '2' {
				az := hub.rotator.Azimuth()
				el := hub.rotator.Elevation()
				if err := c.write(fmt.Sprintf("+0%.3d+0%.3d\r\n", az, el)); err != nil {
					log.Println(err)
					return
				}
				// only azimuth
			} else {
				az := hub.rotator.Azimuth()
				if err := c.write(fmt.Sprintf("+0%.3d\r\n", az)); err != nil {
					log.Println(err)
					return
				}
			}
		// stop azimuth
		case "A":
			if err := hub.rotator.StopAzimuth(); err != nil {
				log.Println(err)
				return
			}
		// stop elevation
		case "E":
			if err := hub.rotator.StopElevation(); err != nil {
				log.Println(err)
				return
			}
		// stop all
		case "S":
			if err := hub.rotator.Stop(); err != nil {
				log.Println(err)
				return
			}
			// unknown commando
		default:
			if err := c.write("?>"); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

// writes a prompt to the tcp socket
func (c *TCPClient) prompt() error {
	if _, err := c.Conn.Write([]byte("?>")); err != nil {
		return err
	}
	return nil
}

// write takes an empty interface and writes it's value to the client's
// tcp socket. If the value in the interface is not supported a log
// message will be printed. If it is not possible to write successfully
// to the socket, an error will be returned with the details.
func (c *TCPClient) write(v interface{}) error {

	data := []byte{}
	switch t := v.(type) {
	case []byte:
		data = v.([]byte)
	case string:
		data = []byte(v.(string))
	default:
		log.Printf("no handler for type %v (msg: %v)\n", t, v)
		return nil
	}

	if _, err := c.Conn.Write(data); err != nil {
		return fmt.Errorf("socket write error (%v): %v", c.Conn.RemoteAddr(), err)
	}

	return nil
}
