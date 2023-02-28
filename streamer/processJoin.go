package streamer

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gofrs/uuid"
)

type Join struct {
	Name string `json:"name"`
}

type AcceptJoin struct{}

func processJoin(s *streamer, clientId uuid.UUID, args map[string]interface{}) error {
	if _, ok := args["name"]; !ok {
		return fmt.Errorf("name is required\n")
	}
	name, ok := args["name"].(string)
	if !ok {
		return fmt.Errorf("invalid type for name\n")
	}

	users[clientId] = &user{
		id:               clientId,
		name:             name,
		y:                float32(rand.Intn(MAP_HEIGHT)),
		x:                float32(rand.Intn(MAP_WIDTH)),
		vy:               0,
		vx:               0,
		leftClickLength:  0,
		rightClickLength: 0,
		input:            make(chan Input, 10),
		previnput: Input{
			W:     false,
			A:     false,
			S:     false,
			D:     false,
			Left:  false,
			Right: false,
			Dx:    0,
			Dy:    0,
		},
	}

	var res AcceptJoin = AcceptJoin{}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}
	s.sendTo(clientId, resJSON)

	return nil
}
