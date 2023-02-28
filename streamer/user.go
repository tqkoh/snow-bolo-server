package streamer

import "github.com/gofrs/uuid"

type user struct {
	id               uuid.UUID
	name             string
	y                int
	x                int
	vy               int
	vx               int
	leftClickLength  int
	rightClickLength int
}

var users map[uuid.UUID]*user
