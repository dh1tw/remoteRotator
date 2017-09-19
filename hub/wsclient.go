package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/gorilla/websocket"
)

//WsClient is a wrapper for clients connected through a Websocket
type WsClient struct {
	*websocket.Conn
}

func (c *WsClient) listen(r rotator.Rotator, closer chan<- *WsClient) {
	defer func() {
		closer <- c
	}()

	for {

		messageType, p, err := c.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "(going away)") ||
				strings.Contains(err.Error(), "(abnormal closure)") {
				return
			}
			log.Printf("websocket read error (%v): %v\n", c.Conn.RemoteAddr(), err)
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		req := rotator.Request{}

		if err := json.Unmarshal(p, &req); err != nil {
			log.Printf("websocket json unmarshal error (%v): %v\n", c.RemoteAddr(), err)
		}

		if err := r.ExecuteRequest(req); err != nil {
			log.Printf("websocket unable to execute request (%v): %v\n", c.RemoteAddr(), err)
			return
		}
	}
}

func (c *WsClient) write(v interface{}) error {

	switch t := v.(type) {
	case rotator.Status:
		s := []rotator.Status{v.(rotator.Status)}
		b, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("unable to serialize msg %v: %v", v, err)
		}
		if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
			return err
		}
	default:
		log.Printf("no handler for type %v (msg: %v)\n", t, v)
		return nil
	}

	return nil
}
