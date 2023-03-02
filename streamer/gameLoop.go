package streamer

import (
	"encoding/json"
	"fmt"
	"time"
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
			if input.W == input.S {
				u.Vy += -u.Vy * V_K
			} else if input.W {
				u.Vy += (-u.Vy/V_MAX - 1) * V_MAX * V_K
			} else {
				u.Vy += (1 - u.Vy/V_MAX) * V_MAX * V_K
			}
			if input.A == input.D {
				u.Vx += -u.Vx * V_K
			} else if input.A {
				u.Vx += (-u.Vx/V_MAX - 1) * V_MAX * V_K
			} else {
				u.Vx += (1 - u.Vx/V_MAX) * V_MAX * V_K
			}

			// update position
			u.Y += u.Vy
			u.X += u.Vx

			// update previnput
			u.PrevInput = input

			// update leftClickLength
			if input.Left {
				u.LeftClickLength++
			} else {
				u.LeftClickLength = 0
			}

			// update rightClickLength
			if input.Right {
				u.RightClickLength++
			} else {
				u.RightClickLength = 0
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
			fmt.Printf("update sent\n")
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
