package streamer

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"
	"github.com/downflux/go-kd/point"
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

// kd
var _ point.P = &P{}

type P struct {
	p   vector.V
	tag string
}

func (p *P) P() vector.V { return p.p }

var kdEntities *kd.KD[*P]

func radiusFromMass(mass float64) float64 {
	var r6 = math.Sqrt(6)
	if mass > 2000./9.*r6 {
		return (-math.Pow(math.E, -(mass-2000./9.*r6)/10000)+1)*RADIUS_M +
			10./3.*r6
	}
	return math.Pow(mass, 1./3.)
}

func processCollide(t *user, u *user) {
	// check if t and u are approaching each other
	var dy = t.Y - u.Y
	var dx = t.X - u.X
	var dl = math.Sqrt(dy*dy + dx*dx)
	if dl == 0 {
		dl = math.Nextafter(0, 1)
	}
	var m = t.Mass
	var M = u.Mass
	{
		var dyd = t.Y + t.Vy*CHECK_APPROACHING_EPS - (u.Y + u.Vy*CHECK_APPROACHING_EPS)
		var dxd = t.X + t.Vx*CHECK_APPROACHING_EPS - (u.X + u.Vx*CHECK_APPROACHING_EPS)
		var dld = math.Sqrt(dyd*dyd + dxd*dxd)
		if dld > dl {
			// update position
			var dd = radiusFromMass(m) + radiusFromMass(M) - dl + 1
			t.Y = t.Y + dd*(m/(m+M))*dy/dl
			t.X = t.X + dd*(m/(m+M))*dx/dl
			u.Y = u.Y - dd*(M/(m+M))*dy/dl
			u.X = u.X - dd*(M/(m+M))*dx/dl

			return
		}
	}

	// update velocities of t and u
	var theta = math.Atan2(dy, dx) + math.Pi/2
	var v0ty = t.Vy*math.Cos(theta) - t.Vx*math.Sin(theta)
	var v0tx = t.Vy*math.Sin(theta) + t.Vx*math.Cos(theta)
	var v0uy = u.Vy*math.Cos(theta) - u.Vx*math.Sin(theta)
	var v0ux = u.Vy*math.Sin(theta) + u.Vx*math.Cos(theta)
	var v0y = v0ty - v0uy
	var v0x = v0tx - v0ux
	var e = 1.
	var v1ty = (v0y*(m-e*M))/(m+M) + v0uy
	var v1uy = (v0y*(m+e*m))/(m+M) + v0uy
	var v1tx = v0x + v0ux
	var v1ux = 0 + v0ux
	t.Vy = v1ty*math.Cos(-theta) - v1tx*math.Sin(-theta)
	t.Vx = v1ty*math.Sin(-theta) + v1tx*math.Cos(-theta)
	u.Vy = v1uy*math.Cos(-theta) - v1ux*math.Sin(-theta)
	u.Vx = v1uy*math.Sin(-theta) + v1ux*math.Cos(-theta)

	// update position to avoid double collision
	// var dd = radiusFromMass(m) + radiusFromMass(M) - dl + 1
	// t.Y = t.Y + dd*(m/(m+M))*dy/dl
	// t.X = t.X + dd*(m/(m+M))*dx/dl
	// u.Y = u.Y + dd*(-M/(m+M))*dy/dl
	// u.X = u.X + dd*(-M/(m+M))*dx/dl

	// update hitstop and inoperable frames
	var imt = math.Abs(m*v0ty - m*v1ty)
	var imu = math.Abs(M*v0uy - M*v1uy)
	t.HitStop = int(M / (M + m) * math.Max(0, math.Log(imt)) * HITSTOP_K)
	u.HitStop = int(m / (M + m) * math.Max(0, math.Log(imu)) * HITSTOP_K)
	t.InOperable = t.HitStop * 2
	u.InOperable = u.HitStop * 2
}

func gameLoop(s *streamer) {
	var frame int = 0
	var prev = time.Now()

	for {
		frame += 1

		kdEntities = kd.New(kd.O[*P]{
			Data: []*P{},
			K:    2,
			N:    1,
		})

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
			if u.InOperable > 0 {
				u.InOperable -= 1
				// input.W = false
				// input.A = false
				// input.S = false
				// input.D = false
				// input.Left = false
				// input.Right = false
			} else {
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
					if l == 0 {
						l = math.Nextafter(0, 1)
					}
					var t = (u.Vx*float64(u.Dx) + u.Vy*float64(u.Dy)) / (l * l)
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
						Vy:    Hy + float64(u.Dy)/l*BULLET_V,
						Vx:    Hx + float64(u.Dx)/l*BULLET_V,
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

			kdEntities.Insert(&P{
				p:   vector.V{u.X, u.Y},
				tag: u.Id.String() + "U",
			})
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
			kdEntities.Insert(&P{
				p:   vector.V{b.Y, b.X},
				tag: b.Id.String() + "B",
			})
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

			kdEntities.Insert(&P{
				p:   vector.V{f.Y, f.X},
				tag: f.Id.String() + "F",
			})
		}

		// process collision
		for _, u := range users {
			for _, p := range kd.KNN(kdEntities, vector.V{u.Y, u.X}, 5, func(p *P) bool { return true }) {
				var id, err = uuid.FromString(p.tag[:len(p.tag)-1])
				if err != nil {
					fmt.Printf("uuid.FromString: p.tag was %v", p.tag)
					panic(err)
				}

				if p.tag[len(p.tag)-1] == 'U' {
					var other = users[id]

					var dy = u.Y - other.Y
					var dx = u.X - other.X
					var l = math.Sqrt(dy*dy + dx*dx)
					if l <= radiusFromMass(u.Mass)+radiusFromMass(other.Mass) && u.Id != other.Id && u.InOperable == 0 && other.InOperable == 0 {
						// collision
						processCollide(u, other)
					}
				} else if p.tag[len(p.tag)-1] == 'B' {
					var other = bullets[id]

					var dy = u.Y - other.Y
					var dx = u.X - other.X
					var l = math.Sqrt(dy*dy + dx*dx)
					if l <= radiusFromMass(u.Mass)+radiusFromMass(other.Mass) && u.Id != other.Owner {
						// collision
						u.Strength -= int(other.Mass/u.Mass) * 50
						u.Mass += other.Mass * BULLET_K
						u.Vy += other.Vy
						delete(bullets, id)
						kdEntities.Remove(p.p, func(q *P) bool { return p.tag == q.tag })
					}
				} else if p.tag[len(p.tag)-1] == 'F' {
					var other = feeds[id]

					var dy = u.Y - other.Y
					var dx = u.X - other.X
					var l = math.Sqrt(dy*dy + dx*dx)
					if l <= radiusFromMass(u.Mass)+radiusFromMass(other.Mass) {
						// collision
						u.Mass += other.Mass
						delete(feeds, id)
						kdEntities.Remove(p.p, func(q *P) bool { return p.tag == q.tag })
					}
				}
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
					HitStop:          user.HitStop,
					InOperable:       user.InOperable,
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
