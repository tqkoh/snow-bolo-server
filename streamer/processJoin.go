package streamer

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/downflux/go-geometry/nd/vector"
	"github.com/gofrs/uuid"
)

type Join struct {
	Name string `json:"name"`
}

type JoinArgs struct {
	Id uuid.UUID `json:"id"`
}

type JoinAccepted struct {
	Method string   `json:"method"`
	Args   JoinArgs `json:"args"`
}

func processJoin(s *streamer, clientId uuid.UUID, args map[string]interface{}) error {
	if _, ok := args["name"]; !ok {
		return fmt.Errorf("name is required\n")
	}
	name, ok := args["name"].(string)
	if !ok {
		return fmt.Errorf("invalid type for name\n")
	}

	var y = float64(rand.Intn(MAP_HEIGHT-MAP_MARGIN*2) + MAP_MARGIN)
	var x = float64(rand.Intn(MAP_WIDTH-MAP_MARGIN*2) + MAP_MARGIN)

	users[clientId] = &user{
		Id:        clientId,
		Dummy:     false,
		Name:      name,
		Mass:      MASS_INIT,
		Strength:  STRENGTH_INIT,
		Enemy:     uuid.Nil,
		Y:         y,
		X:         x,
		Input:     make(chan Input, 10),
		PrevInput: Input{},
	}
	kdEntities.Insert(&P{
		p:   vector.V{y, x},
		tag: clientId.String() + "U",
	})

	var res JoinAccepted = JoinAccepted{
		Method: "joinAccepted",
		Args: JoinArgs{
			Id: clientId,
		},
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}
	s.sendTo(clientId, resJSON)

	return nil
}
