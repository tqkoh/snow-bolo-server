package game

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/tqkoh/snow-bolo-server/streamer"
)

func processActive(s *streamer.Streamer, clientId uuid.UUID, args map[string]interface{}) error {
	active, ok := args["active"].(bool)
	if !ok {
		return fmt.Errorf("invalid type for active\n")
	}

	c, ok := s.Clients[clientId]
	if !ok {
		println("client not found")
		return nil
	}
	c.Active = active

	return nil
}
