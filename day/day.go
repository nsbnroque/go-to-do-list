package day

import (
	"time"

	"github.com/nsbnroque/go-to-do-list/task"
)

type Day struct {
	weekday time.Weekday
	tasks   []task.Task
}
