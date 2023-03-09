package streamer

import (
	"fmt"

	"github.com/gofrs/uuid"
)

func processActive(s *streamer, clientId uuid.UUID, args map[string]interface{}) error {
	active, ok := args["active"].(bool)
	if !ok {
		return fmt.Errorf("invalid type for active\n")
	}

	c, ok := s.clients[clientId]
	if !ok {
		println("client not found")
		return nil
	}
	c.active = active

	return nil
}
