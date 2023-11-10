package syncchannel

import (
	h "github.com/nsbnroque/go-to-do-list/home"
	t "github.com/nsbnroque/go-to-do-list/task"
	u "github.com/nsbnroque/go-to-do-list/user"
)

func SyncTasks(home *h.Home, syncChannel SyncChannel) {
	for {
		select {
		case completedTask := <-syncChannel.CompleteTask:
			// Lógica para atualizar pontuação do morador
			updateRanking(completedTask.User, completedTask.Task)
			//u.UpdateScore(completedTask.User.Email, completedTask.Task.Name)
		}
	}
}

func updateRanking(user u.User, task t.Task) {
	//TO-DO
}

func completeTask(task t.Task, user u.User, syncChannel SyncChannel) {
	// Lógica para marcar tarefa como concluída no banco de dados
	completedTask := CompletedTask{
		Task: task,
		User: user,
	}
	syncChannel.CompleteTask <- completedTask
}
