package game

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/downflux/go-geometry/nd/vector"
	"github.com/gofrs/uuid"
	"github.com/tqkoh/snow-bolo-server/streamer"
)

type Join struct {
	Name string `json:"name"`
}

func processJoin(s *streamer.Streamer, clientId uuid.UUID, args map[string]interface{}) error {
	if _, ok := args["name"]; !ok {
		return fmt.Errorf("name is required")
	}
	name, ok := args["name"].(string)
	if !ok {
		return fmt.Errorf("invalid type for name")
	}

	var message = fmt.Sprintf("%s joined", name)
	var p = streamer.Payload{
		Method: "message",
		Args: map[string]interface{}{
			"message": message,
		},
	}
	println("chat: ", message)

	resJSON, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	s.Send(resJSON, func(_ *streamer.Client) bool { return true })

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

	var res = streamer.Payload{
		Method: "joinAccepted",
		Args: map[string]interface{}{
			"id": clientId,
		},
	}
	resJSON, err = json.Marshal(res)
	if err != nil {
		return err
	}
	s.SendTo(clientId, resJSON)

	return nil
}
