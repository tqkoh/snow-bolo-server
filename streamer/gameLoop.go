package streamer

import "time"

const V_MAX = 2
const V_K = 0.5

func gameLoop() {
	var prev = time.Now()
	for {
		// user
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
		// wait for next frame
		var now = time.Now()
		var next = prev.Add(time.Second / 60)
		if next.After(now) {
			time.Sleep(next.Sub(now))
		}
		prev = next
	}
}
