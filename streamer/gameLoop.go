package streamer

import (
	"encoding/json"
	"math"
	"time"

	"github.com/gofrs/uuid"
)

type updateArgs struct {
	Users   []userReduced `json:"users"`
	Bullets []bullet      `json:"bullets"`
	Feeds   []feed        `json:"feeds"`
}

type update struct {
	Method string     `json:"method"`
	Args   updateArgs `json:"args"`
}

func gameLoop(s *streamer) {
	var frame int = 0
	var prev = time.Now()
	for {
		frame += 1
		// process users' input and update state
		for _, u := range users {
			// process input
			var input Input
			for len(u.Input) > 5 {
				<-u.Input
			}
			if len(u.Input) == 0 {
				input = u.PrevInput
			} else {
				input = <-u.Input
			}

			// update velocity
			var vLimit = V_MIN
			if u.Mass < V_ATTACK {
				vLimit += (V_ATTACK - u.Mass) / V_ATTACK * (V_MAX - V_MIN)
			}

			if input.W == input.S {
				u.Vy += -u.Vy * V_K
			} else if input.W {
				u.Vy += (-u.Vy/vLimit - 1) * vLimit * V_K
			} else {
				u.Vy += (1 - u.Vy/vLimit) * vLimit * V_K
			}
			if input.A == input.D {
				u.Vx += -u.Vx * V_K
			} else if input.A {
				u.Vx += (-u.Vx/vLimit - 1) * vLimit * V_K
			} else {
				u.Vx += (1 - u.Vx/vLimit) * vLimit * V_K
			}

			// update position
			u.Y += u.Vy
			u.X += u.Vx

			if u.Y < MAP_MARGIN {
				u.Y = MAP_MARGIN
				u.Vy = 0
			}
			if u.Y >= MAP_HEIGHT-MAP_MARGIN {
				u.Y = MAP_HEIGHT - MAP_MARGIN
				u.Vy = 0
			}
			if u.X < MAP_MARGIN {
				u.X = MAP_MARGIN
				u.Vx = 0
			}
			if u.X >= MAP_HEIGHT-MAP_MARGIN {
				u.X = MAP_HEIGHT - MAP_MARGIN
				u.Vx = 0
			}

			u.Mass += math.Sqrt(u.Vy*u.Vy+u.Vx*u.Vx) * math.Sqrt(u.Mass) * MASS_K

			u.Dy = input.Dy
			u.Dx = input.Dx

			// update previnput
			u.PrevInput = input

			// update leftClickLength
			if input.Left {
				if u.LeftClickLength < 60 {
					u.LeftClickLength++
				}
			} else {
				if u.LeftClickLength > 0 {
					var id = uuid.Must(uuid.NewV4())
					var l = math.Sqrt(float64(u.Dy*u.Dy + u.Dx*u.Dx))
					var t = l*l - (u.Vx*float64(u.Dx) + u.Vy*float64(u.Dy))
					var Hx = float64(u.Dx) * t
					var Hy = float64(u.Dy) * t
					var mass = u.Mass * float64(u.LeftClickLength) / 60 * MAX_BULLET_MASS
					bullets[id] = &bullet{
						Id:    id,
						Owner: u.Id,
						Mass:  mass,
						Life:  BULLET_LIFE,
						Y:     u.Y,
						X:     u.X,
						Vy:    Hy - Hy + float64(u.Dy)/l*BULLET_V,
						Vx:    Hx - Hx + float64(u.Dx)/l*BULLET_V,
					}

					u.Mass -= mass
				}
				u.LeftClickLength = 0
			}

			// update rightClickLength
			if input.Right {
				u.RightClickLength++
			} else {
				u.RightClickLength = 0
			}
		}

		// update bullets' state
		for _, b := range bullets {
			b.Life -= 1

			if b.Life <= 0 {
				feeds[b.Id] = &feed{
					Id:   b.Id,
					Mass: b.Mass,
					Y:    b.Y,
					X:    b.X,
					Vy:   b.Vy,
					Vx:   b.Vx,
				}
				delete(bullets, b.Id) // safe
				continue
			}

			b.Y += b.Vy
			b.X += b.Vx
			if b.Y < MAP_MARGIN {
				b.Y = MAP_MARGIN
				b.Vy = 0
			}
			if b.Y >= MAP_HEIGHT-MAP_MARGIN {
				b.Y = MAP_HEIGHT - MAP_MARGIN
				b.Vy = 0
			}
			if b.X < MAP_MARGIN {
				b.X = MAP_MARGIN
				b.Vx = 0
			}
			if b.X >= MAP_HEIGHT-MAP_MARGIN {
				b.X = MAP_HEIGHT - MAP_MARGIN
				b.Vx = 0
			}
		}

		// update feeds' state
		for _, f := range feeds {
			f.Vy += -f.Vy * V_K
			f.Vx += -f.Vx * V_K

			f.Y += f.Vy
			f.X += f.Vx
			if f.Y < MAP_MARGIN {
				f.Y = MAP_MARGIN
				f.Vy = 0
			}
			if f.Y >= MAP_HEIGHT-MAP_MARGIN {
				f.Y = MAP_HEIGHT - MAP_MARGIN
				f.Vy = 0
			}
			if f.X < MAP_MARGIN {
				f.X = MAP_MARGIN
				f.Vx = 0
			}
			if f.X >= MAP_HEIGHT-MAP_MARGIN {
				f.X = MAP_HEIGHT - MAP_MARGIN
				f.Vx = 0
			}
		}

		// send state to all clients
		if frame%SEND_STATE_PER == 0 {
			var u []userReduced = make([]userReduced, 0)
			for _, user := range users {
				u = append(u, userReduced{
					Id:               user.Id,
					Name:             user.Name,
					Mass:             user.Mass,
					Strength:         user.Strength,
					Y:                user.Y,
					X:                user.X,
					Vy:               user.Vy,
					Vx:               user.Vx,
					Dy:               user.Dy,
					Dx:               user.Dx,
					LeftClickLength:  user.LeftClickLength,
					RightClickLength: user.RightClickLength,
				})
			}
			var b []bullet = make([]bullet, 0)
			for _, bullet := range bullets { // todo: send only bullets or feed when appear
				b = append(b, *bullet)
			}
			var f []feed = make([]feed, 0)
			for _, feed := range feeds {
				f = append(f, *feed)
			}
			var args = updateArgs{
				Users:   u,
				Bullets: b,
				Feeds:   f,
			}
			var update = update{
				Method: "update",
				Args:   args,
			}
			var updateJSON, _ = json.Marshal(update)
			s.send(updateJSON, func(c *client) bool { return true })
		}

		// wait for next frame
		var now = time.Now()
		var next = prev.Add(time.Second / 60)
		if next.After(now) {
			time.Sleep(next.Sub(now))
		}
		prev = next
	}
}
