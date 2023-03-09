package streamer

import (
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type receiveData struct {
	id      uuid.UUID
	payload []byte
}

type client struct {
	id       uuid.UUID
	conn     *websocket.Conn
	receiver chan receiveData
	sender   chan []byte
	closer   chan bool
	active   bool
}

func newClient(roomID string, conn *websocket.Conn, receiver chan receiveData) *client {
	return &client{
		id:       uuid.Must(uuid.NewV4()),
		conn:     conn,
		receiver: receiver,
		sender:   make(chan []byte),
		closer:   make(chan bool),
		active:   true,
	}
}
func (c *client) listen() {
	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			c.closer <- true
			return
		}
		if messageType != websocket.TextMessage {
			continue
		}
		// fmt.Printf("message: %s\n", message)

		c.receiver <- receiveData{
			id:      c.id,
			payload: message,
		}
	}
}

func (c *client) send() {
	for {
		message, ok := <-c.sender

		if !ok {
			c.closer <- true
			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, message)

		if err != nil {
			c.closer <- true
			return
		}
		// fmt.Printf("sent: %s\n", message)
	}
}
