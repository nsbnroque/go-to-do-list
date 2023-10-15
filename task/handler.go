package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func CreateTaskHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userEmail := r.URL.Query().Get("user")

		var taskData Task
		err := json.NewDecoder(r.Body).Decode(&taskData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err), http.StatusBadRequest)
			log.Println(err.Error())
			return
		}
		// Execute query
		result, err := neo4j.ExecuteQuery(ctx, driver,
			`MATCH (u:User {email: $email})
			MERGE (t:Task {name: $name})
				SET t.reward = $reward
			MERGE (u)-[r:HAS_TASK]->(t)
			ON MATCH SET r.status = $status
			RETURN t.name as name, t.reward as reward, r.status as status;			
			`,
			map[string]interface{}{
				"name":   taskData.Name,
				"status": "pending",
				"email":  userEmail,
				"reward": taskData.Reward,
			},
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao criar task: %v", err), http.StatusInternalServerError)
			return
		}

		for _, record := range result.Records {
			name, _ := record.Get("name")
			reward, _ := record.Get("reward")
			task := Task{
				Name:   name.(string),
				Reward: reward.(int64),
			}
			// Enviar a lista de usuários como resposta
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)

		}
	}
}

func ChangeTaskHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userEmail := r.URL.Query().Get("user")

		var taskData Task
		err := json.NewDecoder(r.Body).Decode(&taskData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err), http.StatusBadRequest)
			log.Println(err.Error())
			return
		}

		// Execute query
		result, err := neo4j.ExecuteQuery(ctx, driver,
			`MATCH (u:User {email: $email})
			MATCH (t:Task {name: $name})
				SET t.reward = $reward
			MATCH (u)-[r:HAS_TASK]->(t)
				SET r.status = $status
			RETURN t.name as name, t.reward as reward, r.status as status;			
			`,
			map[string]interface{}{
				"name":   taskData.Name,
				"status": taskData.Status,
				"reward": taskData.Reward,
				"email":  userEmail,
			},
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao criar task: %v", err), http.StatusInternalServerError)
			return
		}

		for _, record := range result.Records {
			name, _ := record.Get("name")
			status, _ := record.Get("status")
			reward, _ := record.Get("reward")
			taskStatus := Status(status.(string))

			task := Task{
				Name:   name.(string),
				Reward: reward.(int64),
				Status: taskStatus,
			}
			// Enviar a lista de usuários como resposta
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)

		}
	}
}

func DeleteTaskHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userEmail := r.URL.Query().Get("user")
		taskName := r.URL.Query().Get("task")

		// Execute query
		_, err := neo4j.ExecuteQuery(ctx, driver,
			`MATCH (u:User {email: $email})
			MATCH (t:Task {name: $taskName})
			MATCH (u)-[r:HAS_TASK]->(t)
			DETACH DELETE t;			
			`,
			map[string]interface{}{
				"taskName": taskName,
				"email":    userEmail,
			},
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao excluir a tarefa: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func GetTasksForUserHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userEmail := r.URL.Query().Get("user")

		// Execute query to get tasks for the user
		result, err := neo4j.ExecuteQuery(ctx, driver,
			`MATCH (u:User {email: $email})-[:HAS_TASK]->(t:Task)
			RETURN t.name as name, t.reward as reward, t.status as status;`,
			map[string]interface{}{
				"email": userEmail,
			},
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao obter as tarefas do usuário: %v", err), http.StatusInternalServerError)
			return
		}

		var tasks []Task
		for _, record := range result.Records {
			task := Task{}

			// Verificar e atribuir o nome
			if name, found := record.Get("name"); found {
				task.Name = name.(string)
			}

			// Verificar e atribuir a recompensa
			if reward, found := record.Get("reward"); found && reward != nil {
				if rewardInt, ok := reward.(int64); ok {
					task.Reward = rewardInt
				} else {
					task.Reward = 0
				}
			}

			// Verificar e atribuir o status
			if status, found := record.Get("status"); found && status != nil {
				if statusStr, ok := status.(string); ok {
					task.Status = Status(statusStr)
				} else {
					log.Println("Erro ao converter status para string")
				}
			}
			tasks = append(tasks, task)
		}

		// Enviar a lista de tarefas como resposta
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}
