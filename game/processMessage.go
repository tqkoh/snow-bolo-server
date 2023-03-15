package game

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/tqkoh/snow-bolo-server/streamer"
)

func processMessage(s *streamer.Streamer, clientId uuid.UUID, args map[string]interface{}) error {
	message, ok := args["message"].(string)
	if !ok {
		return fmt.Errorf("invalid type for message")
	}

	var res = streamer.Payload{
		Method: "message",
		Args: map[string]interface{}{
			"message": message,
		},
	}
	println("chat: ", message)

	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}

	// s.send([]byte(sendMessage), func(_ *client) bool { return true })
	s.Send(resJSON, func(_ *streamer.Client) bool { return true })
	return nil
}
