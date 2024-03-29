package game

import "github.com/gofrs/uuid"

type user struct {
	Id                 uuid.UUID  `json:"id"`
	Dummy              bool       `json:"dummy"`
	Name               string     `json:"name"`
	Mass               float64    `json:"mass"`
	Strength           float64    `json:"strength"`
	Damage             int        `json:"damage"`
	Enemy              uuid.UUID  `json:"enemy"`
	CombatFrame        int        `json:"combatFrame"`
	Kills              int        `json:"kills"`
	Y                  float64    `json:"y"`
	X                  float64    `json:"x"`
	Vy                 float64    `json:"vy"`
	Vx                 float64    `json:"vx"`
	Dy                 int        `json:"dy"`
	Dx                 int        `json:"dx"`
	LeftClickLength    int        `json:"leftClickLength"`
	RightClickLength   int        `json:"rightClickLength"`
	LeftClickCancelled bool       `json:"leftClickCancelled"`
	HitStop            int        `json:"hitStop"`
	InOperable         int        `json:"inOperable"`
	Input              chan Input `json:"input"`
	PrevInput          Input      `json:"prevInput"`
}

var users map[uuid.UUID]*user = make(map[uuid.UUID]*user)

type userReduced struct {
	Id               uuid.UUID `json:"id"`
	Dummy            bool      `json:"dummy"`
	Name             string    `json:"name"`
	Mass             float64   `json:"mass"`
	Strength         int       `json:"strength"`
	Damage           int       `json:"damage"`
	Y                float64   `json:"y"`
	X                float64   `json:"x"`
	Vy               float64   `json:"vy"`
	Vx               float64   `json:"vx"`
	Dy               int       `json:"dy"`
	Dx               int       `json:"dx"`
	LeftClickLength  int       `json:"leftClickLength"`
	RightClickLength int       `json:"rightClickLength"`
	HitStop          int       `json:"hitStop"`
	InOperable       int       `json:"inOperable"`
}

type bullet struct {
	Id    uuid.UUID `json:"id"`
	Owner uuid.UUID `json:"owner"`
	Mass  float64   `json:"mass"`
	Life  int       `json:"life"`
	Y     float64   `json:"y"`
	X     float64   `json:"x"`
	Vy    float64   `json:"vy"`
	Vx    float64   `json:"vx"`
}

var bullets map[uuid.UUID]*bullet = make(map[uuid.UUID]*bullet)

type feed struct {
	Id   uuid.UUID `json:"id"`
	Mass float64   `json:"mass"`
	Y    float64   `json:"y"`
	X    float64   `json:"x"`
	Vy   float64   `json:"vy"`
	Vx   float64   `json:"vx"`
}

var feeds map[uuid.UUID]*feed = make(map[uuid.UUID]*feed)
