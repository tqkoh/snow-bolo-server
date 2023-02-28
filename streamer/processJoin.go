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

type AcceptJoin struct {
	Id uuid.UUID `json:"id"`
}

func processJoin(s *streamer, args map[string]interface{}) error {
	if _, ok := args["name"]; !ok {
		return fmt.Errorf("name is required\n")
	}
	name, ok := args["name"].(string)
	if !ok {
		return fmt.Errorf("invalid type for name\n")
	}
	id := uuid.Must(uuid.NewV4())

	users[id] = &user{
		id:               id,
		name:             name,
		y:                rand.Intn(MAP_HEIGHT),
		x:                rand.Intn(MAP_WIDTH),
		vy:               0,
		vx:               0,
		leftClickLength:  0,
		rightClickLength: 0,
	}

	var res AcceptJoin = AcceptJoin{
		Id: id,
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}
	s.sendTo(id, resJSON)

	return nil
}
