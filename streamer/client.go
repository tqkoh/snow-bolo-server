package streamer

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type receiveData struct {
	id      uuid.UUID
	roomID  string
	payload []byte
}

type client struct {
	id       uuid.UUID
	roomId   string
	conn     *websocket.Conn
	receiver chan receiveData
	sender   chan string
	closer   chan bool
}

func newClient(roomID string, conn *websocket.Conn, receiver chan receiveData) *client {
	return &client{
		id:       uuid.Must(uuid.NewV4()),
		roomId:   roomID,
		conn:     conn,
		receiver: receiver,
		sender:   make(chan string),
		closer:   make(chan bool),
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
		fmt.Printf("message: %s, roomID: %s\n", message, c.roomId)

		c.receiver <- receiveData{
			id:      c.id,
			roomID:  c.roomId,
			payload: message,
		}
	}
}

func (c *client) send() {
	for {
		message := <-c.sender

		err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))

		if err != nil {
			c.closer <- true
			return
		}
	}
}
