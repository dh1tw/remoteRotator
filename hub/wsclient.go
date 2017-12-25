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

func (c *WsClient) listen(hub *Hub, closer chan<- *WsClient) {
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

		hub.RLock()
		r, ok := hub.rotators[req.Name]
		hub.RUnlock()

		if !ok {
			fmt.Printf("request for unknown rotator %s\n", req.Name)
			continue
		}

		if err := r.ExecuteRequest(req); err != nil {
			log.Printf("websocket unable to execute request (%v): %v\n", c.RemoteAddr(), err)
			r.Close()
		}
	}
}

func (c *WsClient) write(event Event) error {

	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("unable to serialize msg %v: %v", event, err)
	}
	if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
		return err
	}

	// switch t := v.(type) {
	// case rotator.Status:
	// 	s := []rotator.Status{v.(rotator.Status)}
	// 	b, err := json.Marshal(s)
	// 	if err != nil {
	// 		return fmt.Errorf("unable to serialize msg %v: %v", v, err)
	// 	}
	// 	if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
	// 		return err
	// 	}
	// case []rotator.Status:
	// 	b, err := json.Marshal(v.([]rotator.Status))
	// 	if err != nil {
	// 		return fmt.Errorf("unable to serialize msg %v: %v", v, err)
	// 	}
	// 	if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
	// 		return err
	// 	}
	// default:
	// 	log.Printf("no handler for type %v (msg: %v)\n", t, v)
	// 	return nil
	// }

	return nil
}
