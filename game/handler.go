package game

import (
	"encoding/json"
	"fmt"

	"github.com/tqkoh/snow-bolo-server/streamer"
)

func HandlerWebSocket(s *streamer.Streamer, data streamer.ReceiveData) error {
	var req streamer.Payload
	err := json.Unmarshal(data.Payload, &req)
	if err != nil {
		return err
	}

	// fmt.Printf("payload: %v\n", data.payload)
	// fmt.Printf("method: %s\n", req.Method)
	// fmt.Printf("args: %v\n", req.Args)

	switch req.Method {
	case "message":
		// fmt.Printf("message received\n")
		processMessage(s, data.Id, req.Args)
	case "join":
		// fmt.Printf("join received\n")
		processJoin(s, data.Id, req.Args)
	case "input":
		// fmt.Printf("input received\n")
		processInput(s, data.Id, req.Args)
	case "active":
		processActive(s, data.Id, req.Args)
	default:
		fmt.Printf("invalid method")
	}

	return nil
}
