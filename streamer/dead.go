package streamer

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"github.com/downflux/go-geometry/nd/vector"
	"github.com/gofrs/uuid"
	"github.com/tqkoh/snowball-server/streamer/utils"
)

type deadArgs struct {
}
type dead struct {
	Method string   `json:"method"`
	Args   deadArgs `json:"args"`
}

func processDeadDisconnected(s *streamer, uId uuid.UUID) {
	name := "unknown"
	if u, ok := users[uId]; ok {
		name = u.Name
		if u.Enemy != uuid.Nil {
			enemyName := "enemy"
			if enemy, ok := users[u.Enemy]; ok {
				enemyName = enemy.Name
			}
			processDead(s, uId, u.Enemy, fmt.Sprintf("%v destroyed by %v", name, enemyName), true)
		}
	}

	processDead(s, uId, uId, fmt.Sprintf("%v disconnected", name), true)
}

func processDead(s *streamer, uId uuid.UUID, by uuid.UUID, log string, disconnected bool) {
	u, ok := users[uId]
	if !ok {
		return
	}

	var m = BroadcastMessage{
		Method: "message",
		Args: MessageArgs{
			Message: log,
		},
	}
	resJSON, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	s.send(resJSON, func(_ *client) bool { return true })

	if !disconnected {
		var m2 = dead{
			Method: "dead",
			Args:   deadArgs{},
		}
		resJSON2, err := json.Marshal(m2)
		if err != nil {
			panic(err)
		}
		if err = s.sendTo(uId, resJSON2); err != nil {
			fmt.Println("sendTo error: ", err)
		}
	}

	// destroy
	{
		var id = uuid.Must(uuid.NewV4())
		feeds[id] = &feed{
			Id:   id,
			Mass: u.Mass * DEAD_MASS_CENTER,
			Y:    u.Y,
			X:    u.X,
			Vy:   u.Vy,
			Vx:   u.Vx,
		}
	}
	for i := 0; i < DEAD_MASS_MINI_NUM; i++ {
		var id = uuid.Must(uuid.NewV4())
		var theta = rand.Float64() * 2 * math.Pi
		var r = radiusFromMass(u.Mass * DEAD_MASS_CENTER)
		feeds[id] = &feed{
			Id:   id,
			Mass: u.Mass * DEAD_MASS_MINI,
			Y:    u.Y + r*math.Sin(theta),
			X:    u.X + r*math.Cos(theta),
			Vy:   u.Vy + DEAD_MASS_MINI_V*math.Sin(theta),
			Vx:   u.Vx + DEAD_MASS_MINI_V*math.Cos(theta),
		}
	}
	kdEntities.Remove(vector.V{u.Y, u.X}, func(q *P) bool { return uId.String() == q.tag })
	utils.Del(users, uId)
}
