package interfaces

import "time"

type Spam interface {
	Get(level int) time.Time
	Set(level int, t time.Time)
}
