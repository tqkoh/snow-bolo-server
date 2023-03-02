package streamer

import "github.com/gofrs/uuid"

type user struct {
	Id               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Mass             float32    `json:"mass"`
	Strength         int        `json:"strength"`
	Y                float32    `json:"y"`
	X                float32    `json:"x"`
	Vy               float32    `json:"vy"`
	Vx               float32    `json:"vx"`
	LeftClickLength  int        `json:"leftClickLength"`
	RightClickLength int        `json:"rightClickLength"`
	Input            chan Input `json:"input"`
	PrevInput        Input      `json:"prevInput"`
}

var users map[uuid.UUID]*user = make(map[uuid.UUID]*user)

type userReduced struct {
	Id               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Mass             float32   `json:"mass"`
	Strength         int       `json:"strength"`
	Y                float32   `json:"y"`
	X                float32   `json:"x"`
	Vy               float32   `json:"vy"`
	Vx               float32   `json:"vx"`
	LeftClickLength  int       `json:"leftClickLength"`
	RightClickLength int       `json:"rightClickLength"`
}

type bullet struct {
	Id    uuid.UUID `json:"id"`
	Owner uuid.UUID `json:"owner"`
	Mass  float32   `json:"mass"`
	Y     float32   `json:"y"`
	X     float32   `json:"x"`
	Vy    float32   `json:"vy"`
	Vx    float32   `json:"vx"`
}

var bullets map[uuid.UUID]*bullet = make(map[uuid.UUID]*bullet)

type feed struct {
	Id   uuid.UUID `json:"id"`
	Mass float32   `json:"mass"`
	Y    float32   `json:"y"`
	X    float32   `json:"x"`
}

var feeds map[uuid.UUID]*feed = make(map[uuid.UUID]*feed)
