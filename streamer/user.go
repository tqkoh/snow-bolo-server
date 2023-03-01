package streamer

import "github.com/gofrs/uuid"

type user struct {
	id               uuid.UUID
	name             string
	y                float32
	x                float32
	vy               float32
	vx               float32
	leftClickLength  int
	rightClickLength int
	input            chan Input
	previnput        Input
}

var users map[uuid.UUID]*user = make(map[uuid.UUID]*user)
