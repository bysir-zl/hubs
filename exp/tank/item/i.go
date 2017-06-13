package item

import "time"

type Interface interface {
	Update(duration time.Duration)
}
