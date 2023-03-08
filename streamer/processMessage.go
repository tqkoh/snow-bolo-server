package streamer

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
)

type MessageArgs struct {
	Message string `json:"message"`
}

type BroadcastMessage struct {
	Method string      `json:"method"`
	Args   MessageArgs `json:"args"`
}

func processMessage(s *streamer, clientId uuid.UUID, args map[string]interface{}) error {
	sendMessage, ok := args["message"].(string)
	if !ok {
		return fmt.Errorf("invalid type for message\n")
	}

	var res = BroadcastMessage{
		Method: "message",
		Args: MessageArgs{
			Message: sendMessage,
		},
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}

	// s.send([]byte(sendMessage), func(_ *client) bool { return true })
	s.send(resJSON, func(_ *client) bool { return true })
	return nil
}
