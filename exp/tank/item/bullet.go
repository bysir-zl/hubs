package item

import (
	"time"
	"math"
)

type Bullet struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Ang   float64 `json:"ang"`
	Speed float64 `json:"speed"`
}

func (p *Bullet) Update(d time.Duration) {
	ly := math.Sin(p.Ang/2*math.Pi) * p.Speed
	lx := math.Cos(p.Ang/2*math.Pi) * p.Speed

	p.Y += ly
	p.X += lx
}

func (p *Bullet) Set(speed, ang float64) {
	p.Speed = speed
	p.Ang = ang
}

func NewBullet(x, y, ang float64) *Bullet {
	return &Bullet{
		X:   x,
		Y:   y,
		Ang: ang,
	}
}
