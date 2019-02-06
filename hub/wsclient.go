package hub

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

//WsClient is a wrapper for clients connected through a Websocket
type WsClient struct {
	*websocket.Conn
}

// listen on the websocket. Despite that no data is read, this function
// is necessary to reply to incoming ping messages.
func (c *WsClient) listen(closer chan<- *WsClient) {
	defer func() {
		closer <- c
	}()

	for {
		// in case of an error just return and signal closing down of the ws
		if _, _, err := c.ReadMessage(); err != nil {
			return
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

	return nil
}
