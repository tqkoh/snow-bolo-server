package streamer

import (
	"encoding/json"
	"fmt"
)

func (s *streamer) handlerWebSocket(data receiveData) error {
	var req payload
	err := json.Unmarshal(data.payload, &req)
	if err != nil {
		return err
	}

	fmt.Printf("payload: %v\n", data.payload)
	fmt.Printf("method: %s\n", req.Method)
	fmt.Printf("args: %v\n", req.Args)

	switch req.Method {
	case "message":
		fmt.Printf("message received")
		sendMessage, ok := req.Args["message"].(string)
		if !ok {
			return fmt.Errorf("invalid type for message")
		}
		s.sendToAll(sendMessage)
	default:
		fmt.Printf("invalid method")
	}

	return nil
}
