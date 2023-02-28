package streamer

import (
	"log"

	"github.com/gofrs/uuid"
)

type streamer struct {
	clients  map[uuid.UUID]*client
	receiver chan receiveData
}

func NewStreamer() *streamer {
	return &streamer{
		clients:  make(map[uuid.UUID]*client),
		receiver: make(chan receiveData),
	}
}

type payload struct {
	Method string                 `json:"method,omitempty"`
	Args   map[string]interface{} `json:"args,omitempty"`
}

func (s *streamer) Listen() {
	for {
		data := <-s.receiver

		go func() {
			err := s.handlerWebSocket(data)
			if err != nil {
				log.Print("error: ", err)
			}
		}()
	}
}

func (s *streamer) send(message string, cond func(c *client) bool) error {
	for _, c := range s.clients {
		if cond(c) {
			c.sender <- message
		}
	}
	return nil
}
