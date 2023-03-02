package streamer

import "github.com/gofrs/uuid"

type user struct {
	id               uuid.UUID
	name             string
	mass             float32
	strength         int
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

type userReduced struct {
	id               uuid.UUID
	name             string
	mass             float32
	strength         int
	y                float32
	x                float32
	vy               float32
	vx               float32
	leftClickLength  int
	rightClickLength int
}

type bullet struct {
	id    uuid.UUID
	owner uuid.UUID
	mass  float32
	y     float32
	x     float32
	vy    float32
	vx    float32
}

var bullets map[uuid.UUID]*bullet = make(map[uuid.UUID]*bullet)

type feed struct {
	id   uuid.UUID
	mass float32
	y    float32
	x    float32
}

var feeds map[uuid.UUID]*feed = make(map[uuid.UUID]*feed)
