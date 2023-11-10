package syncchannel

import (
	t "github.com/nsbnroque/go-to-do-list/task"
	u "github.com/nsbnroque/go-to-do-list/user"
)

type SyncChannel struct {
	CompleteTask chan CompletedTask
}

type CompletedTask struct {
	Task t.Task
	User u.User
}
