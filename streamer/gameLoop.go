package streamer

import (
	"encoding/json"
	"time"
)

const V_MAX = 2
const V_K = 0.5

type updateArgs struct {
	Users   []userReduced `json:"users"`
	Bullets []bullet      `json:"bullets"`
	Feeds   []feed        `json:"feeds"`
}

type update struct {
	Method string `json:"method"`
	Args   updateArgs
}

func gameLoop(s *streamer) {
	var prev = time.Now()
	for {
		// process users' input and update state
		for _, u := range users {
			// process input
			var input Input
			for len(u.input) > 5 {
				<-u.input
			}
			if len(u.input) == 0 {
				input = u.previnput
			} else {
				input = <-u.input
			}

			// update velocity
			if input.W {
				u.vy -= (V_MAX - u.vy) * V_K
			}
			if input.A {
				u.vx -= (V_MAX - u.vx) * V_K
			}
			if input.S {
				u.vy += (V_MAX - u.vy) * V_K
			}
			if input.D {
				u.vx += (V_MAX - u.vx) * V_K
			}

			// update position
			u.y += u.vy
			u.x += u.vx

			// update previnput
			u.previnput = input

			// update leftClickLength
			if input.Left {
				u.leftClickLength++
			} else {
				u.leftClickLength = 0
			}

			// update rightClickLength
			if input.Right {
				u.rightClickLength++
			} else {
				u.rightClickLength = 0
			}
		}
		// send state to all clients
		var u []userReduced = make([]userReduced, 0)
		for _, user := range users {
			u = append(u, userReduced{
				id:               user.id,
				name:             user.name,
				mass:             user.mass,
				strength:         user.strength,
				y:                user.y,
				x:                user.x,
				vy:               user.vy,
				vx:               user.vx,
				leftClickLength:  user.leftClickLength,
				rightClickLength: user.rightClickLength,
			})
		}
		var bu []bullet = make([]bullet, 0)
		for _, bullet := range bullets { // todo: send only bullets or feed when appear
			bu = append(bu, *bullet)
		}
		var f []feed = make([]feed, 0)
		for _, feed := range feeds {
			f = append(f, *feed)
		}
		var args = updateArgs{
			Users:   u,
			Bullets: bu,
			Feeds:   f,
		}
		var update = update{
			Method: "update",
			Args:   args,
		}
		var updateJSON, _ = json.Marshal(update)
		s.send(updateJSON, func(c *client) bool { return true })

		// wait for next frame
		var now = time.Now()
		var next = prev.Add(time.Second / 60)
		if next.After(now) {
			time.Sleep(next.Sub(now))
		}
		prev = next
	}
}
