package home

import (
	"github.com/google/uuid"
	"github.com/nsbnroque/go-to-do-list/task"
	"github.com/nsbnroque/go-to-do-list/user"
)

type Home struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	Residents []user.User `json:"residents"`
	Tasks     []task.Task `json:"tasks"`
}
