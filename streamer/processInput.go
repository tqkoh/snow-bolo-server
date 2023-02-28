package streamer

import (
	"github.com/gofrs/uuid"
	"github.com/mitchellh/mapstructure"
)

type Input struct {
	W     bool `json:"w"`
	A     bool `json:"a"`
	S     bool `json:"s"`
	D     bool `json:"d"`
	Left  bool `json:"left"`
	Right bool `json:"right"`
	Dx    int  `json:"dx"`
	Dy    int  `json:"dy"`
}

func processInput(s *streamer, clientId uuid.UUID, args map[string]interface{}) error {
	var input Input
	err := mapstructure.Decode(args, &input)
	if err != nil {
		return err
	}

	// TODO: push input and later process it in the game loop every frame
	users[clientId].input <- input
	return nil
}
