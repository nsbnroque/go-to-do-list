package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/nsbnroque/go-to-do-list/internal/database"
)

func CreateTaskHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	userEmail := c.Query("user")

	var taskData Task
	if err := c.ShouldBindJSON(&taskData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err),
		})
		log.Println(err.Error())
		return
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(c.Request.Context(), dbHandler.Driver,
		`MATCH (u:User {email: $email})-[:LIVES_IN]->(h:Home)
		MERGE (t:Task {name: $name})
		SET t.reward = $reward
		MERGE (h)-[r:HAS_TASK]->(t)
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
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao criar task: %v", err),
		})
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
		c.JSON(http.StatusOK, task)
	}
}

func ChangeTaskHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	_, driver := dbHandler.Ctx, dbHandler.Driver
	userEmail := c.Query("user")

	var taskData Task
	if err := c.ShouldBindJSON(&taskData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err),
		})
		log.Println(err.Error())
		return
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(c.Request.Context(), driver,
		`MATCH (u:User {email: $email})
		MATCH (t:Task {name: $name})
			SET t.reward = $reward
		MATCH (u)-[h:LIVES_IN]->(h)-[r:HAS_TASK]->(t)
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
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao criar task: %v", err),
		})
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
		c.JSON(http.StatusOK, task)
	}
}

func DeleteTaskHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	userEmail := c.Query("user")
	taskName := c.Query("task")

	// Execute query
	_, err = neo4j.ExecuteQuery(c.Request.Context(), dbHandler.Driver,
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
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao excluir a tarefa: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Tarefa %s excluída com sucesso!", taskName),
	})
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
