package item

import (
	"time"
	"math"
)

type Tank struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Ang   float64 `json:"ang"`
	Speed float64 `json:"speed"`
}

func (p *Tank) Update(d time.Duration) {
	t := float64(d) / float64(time.Second)
	ly := math.Sin(p.Ang*math.Pi/180) * p.Speed * t
	lx := math.Cos(p.Ang*math.Pi/180) * p.Speed * t

	p.Y += ly
	p.X += lx
}

func (p *Tank) Set(speed, ang float64) {
	p.Speed = speed
	p.Ang = ang
}

func NewTank(x, y, ang float64) *Tank {
	return &Tank{
		X:   x,
		Y:   y,
		Ang: ang,
	}
}
