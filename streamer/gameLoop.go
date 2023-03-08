package streamer

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"
	"github.com/downflux/go-kd/point"
	"github.com/gofrs/uuid"
	"github.com/tqkoh/snowball-server/streamer/utils"
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

func processCollide(s *streamer, t *user, u *user) {
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
	// var imu = math.Abs(M*v0uy - M*v1uy) // equal to imt
	t.HitStop = HITSTOP
	u.HitStop = HITSTOP
	t.InOperable = int(M / (M + m) * math.Max(0, math.Log(imt)) * INOPERABLE_K)
	u.InOperable = int(m / (M + m) * math.Max(0, math.Log(imt)) * INOPERABLE_K)
	t.CombatFrame = COMBAT_FRAME
	u.CombatFrame = COMBAT_FRAME
	t.Enemy = u.Id
	u.Enemy = t.Id

	t.Damage += int(M / (m + M) * imt * STRENGTH_COLLISION_K)
	u.Damage += int(m / (m + M) * imt * STRENGTH_COLLISION_K)
	t.Strength -= M / (m + M) * imt * STRENGTH_COLLISION_K
	u.Strength -= m / (m + M) * imt * STRENGTH_COLLISION_K
	t.Mass *= COLLIDE_K
	u.Mass *= COLLIDE_K
	if t.Mass < 1 {
		t.Mass = 1
	}
	if u.Mass < 1 {
		u.Mass = 1
	}
	var uName = u.Name
	var uId = u.Id
	if u.Strength <= 0 {
		processDead(s, u.Id, t.Id, fmt.Sprintf("%v was hit by %v", u.Name, t.Name), false)
	}
	if t.Strength <= 0 {
		processDead(s, t.Id, uId, fmt.Sprintf("%v was hit by %v", t.Name, uName), false)
	}
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

				if u.RightClickLength > 0 {

					u.Vy *= PRESS_V_K
					u.Vx *= PRESS_V_K
					u.Strength = float64(u.Strength + PRESS_RECOVER)
					u.Damage = -1
					if u.Strength > 100 {
						u.Strength = 100
					} else {
						u.Mass *= PRESS_REDUCE
						u.Mass -= PRESS_REDUCE_C
						if u.Mass < 1 {
							u.Mass = 1
						}
					}
				}
			}

			// update position
			if u.HitStop > 0 {
				u.HitStop -= 1
				u.Y += float64(rand.Intn(3) - 1)
				u.X += float64(rand.Intn(3) - 1)
			} else {
				u.Y += u.Vy
				u.X += u.Vx
			}
			var radius = radiusFromMass(u.Mass)
			if u.Y < MAP_MARGIN+radius {
				u.Y = MAP_MARGIN + radius
				u.Vy = 0
			}
			if u.Y >= MAP_HEIGHT-MAP_MARGIN-radius {
				u.Y = MAP_HEIGHT - MAP_MARGIN - radius
				u.Vy = 0
			}
			if u.X < MAP_MARGIN+radius {
				u.X = MAP_MARGIN + radius
				u.Vx = 0
			}
			if u.X >= MAP_HEIGHT-MAP_MARGIN-radius {
				u.X = MAP_HEIGHT - MAP_MARGIN - radius
				u.Vx = 0
			}

			if u.Mass < 1 {
				u.Mass = 1
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

					var radiusAfter = radiusFromMass(u.Mass - mass)
					bullets[id] = &bullet{
						Id:    id,
						Owner: u.Id,
						Mass:  mass,
						Life:  BULLET_LIFE,
						Y:     u.Y + radiusAfter*float64(u.Dy)/l,
						X:     u.X + radiusAfter*float64(u.Dx)/l,
						Vy:    Hy + BULLET_V*float64(u.Dy)/l,
						Vx:    Hx + BULLET_V*float64(u.Dx)/l,
					}

					u.Mass -= mass * BULLET_NEED
					if u.Mass < 1 {
						u.Mass = 1
					}
				}
				u.LeftClickLength = 0
			}

			// update rightClickLength
			if input.Right {
				u.RightClickLength++
			} else {
				u.RightClickLength = 0
			}

			u.CombatFrame -= 1
			if u.CombatFrame <= 0 {
				u.Enemy = uuid.Nil
			}

			kdEntities.Insert(&P{
				p:   vector.V{u.Y, u.X},
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
				utils.Del(bullets, b.Id) // safe
				continue
			}

			b.Y += b.Vy
			b.X += b.Vx
			var radius = radiusFromMass(b.Mass)
			if b.Y < MAP_MARGIN+radius {
				b.Y = MAP_MARGIN + radius
				b.Vy = 0
			}
			if b.Y >= MAP_HEIGHT-MAP_MARGIN-radius {
				b.Y = MAP_HEIGHT - MAP_MARGIN - radius
				b.Vy = 0
			}
			if b.X < MAP_MARGIN+radius {
				b.X = MAP_MARGIN + radius
				b.Vx = 0
			}
			if b.X >= MAP_HEIGHT-MAP_MARGIN-radius {
				b.X = MAP_HEIGHT - MAP_MARGIN - radius
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

			radius := radiusFromMass(f.Mass)
			if f.Y < MAP_MARGIN-radius {
				f.Y = MAP_MARGIN - radius
				f.Vy = 0
			}
			if f.Y >= MAP_HEIGHT-MAP_MARGIN+radius {
				f.Y = MAP_HEIGHT - MAP_MARGIN + radius
				f.Vy = 0
			}
			if f.X < MAP_MARGIN-radius {
				f.X = MAP_MARGIN - radius
				f.Vx = 0
			}
			if f.X >= MAP_HEIGHT-MAP_MARGIN+radius {
				f.X = MAP_HEIGHT - MAP_MARGIN + radius
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
					var other, ok = users[id]
					if !ok {
						continue
					}

					var dy = u.Y - other.Y
					var dx = u.X - other.X
					var l = math.Sqrt(dy*dy + dx*dx)
					if l <= radiusFromMass(u.Mass)+radiusFromMass(other.Mass) && u.Id != other.Id && u.HitStop <= 0 && other.HitStop <= 0 {
						// collision
						processCollide(s, u, other)
					}
				} else if p.tag[len(p.tag)-1] == 'B' {
					var other, ok = bullets[id]
					if !ok {
						continue
					}

					var dy = u.Y - other.Y
					var dx = u.X - other.X
					var l = math.Sqrt(dy*dy + dx*dx)
					if l <= radiusFromMass(u.Mass)+radiusFromMass(other.Mass) && u.Id != other.Owner {
						// collision
						var m = u.Mass
						var M = other.Mass
						var v0y = u.Vy
						var v0x = u.Vx
						u.Vy = (u.Vy*m+other.Vy*M*KB_GAIN)/(m+M) + other.Vy*KB_C
						u.Vx = (u.Vx*m+other.Vx*M*KB_GAIN)/(m+M) + other.Vx*KB_C
						var imy = math.Abs(m*v0y - m*u.Vy)
						var imx = math.Abs(m*v0x - m*u.Vx)
						var im = math.Sqrt(imy*imy + imx*imx)
						u.InOperable = INOPERABLE // int(M / (M + m) * math.Max(0, math.Log(im)) * INOPERABLE_K)
						u.Damage += int(M / (m + M) * im * STRENGTH_HIT_K)
						u.Strength -= M / (m + M) * im * STRENGTH_HIT_K
						u.Mass += other.Mass * BULLET_K
						u.Enemy = other.Owner
						u.CombatFrame = COMBAT_FRAME

						if u.Strength <= 0 {
							processDead(s, u.Id, other.Id, fmt.Sprintf("%v was shot by %v", u.Name, users[other.Owner].Name), false)
						}

						utils.Del(bullets, id)
						kdEntities.Remove(p.p, func(q *P) bool { return p.tag == q.tag })
					}
				} else if p.tag[len(p.tag)-1] == 'F' {
					var other, ok = feeds[id]
					if !ok {
						continue
					}

					var dy = u.Y - other.Y
					var dx = u.X - other.X
					var l = math.Sqrt(dy*dy + dx*dx)
					if l <= radiusFromMass(u.Mass)+radiusFromMass(other.Mass) {
						// collision
						u.Mass += other.Mass
						utils.Del(feeds, id)
						kdEntities.Remove(p.p, func(q *P) bool { return p.tag == q.tag })
					}
				}
				// fmt.Printf("%v(%d,%d) | ", p.tag[len(p.tag)-1:], int(p.p[0]), int(p.p[1]))
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
					Strength:         int(math.Min(user.Strength+1, 100)),
					Damage:           user.Damage,
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

				user.Damage = 0
			}
			sort.Slice(u, func(i, j int) bool { return u[i].Mass > u[j].Mass })

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
			var updateJSON, err = json.Marshal(update)
			if err != nil {
				fmt.Printf("args: %v", args)
				fmt.Printf("json.Marshal error: %v", err)
			}
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
