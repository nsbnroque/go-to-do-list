package routes

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nsbnroque/go-to-do-list/home"
	"github.com/nsbnroque/go-to-do-list/internal/database"
	"github.com/nsbnroque/go-to-do-list/task"
	"github.com/nsbnroque/go-to-do-list/user"
)

func HandleRequests() {
	dbHandler, err := database.NewDatabaseHandler()
	if err != nil {
		log.Fatalf("Falha ao obter o handler do banco de dados: %v", err)
	}

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PATCH", "DELETE"}
	r.Use(cors.New(config))

	r.POST("/users", user.CreateUserHandler)
	r.GET("/users", user.FindAllUsersHandler)
	r.GET("/users/find", user.FindByEmailHandler)
	r.PUT("/users", user.UpdateUserHandler)
	r.POST("/tasks", task.CreateTaskHandler)
	r.PUT("/tasks", task.ChangeTaskHandler)
	r.DELETE("/tasks", task.DeleteTaskHandler)
	r.GET("/tasks", func(c *gin.Context) {
		task.GetTasksForUserHandler(dbHandler.Ctx, dbHandler.Driver, dbHandler.Config.Database)(c.Writer, c.Request)
	})
	r.POST("/home", home.CreateHomeHandler)
	r.PATCH("/home", home.AddResidentToHomeHandler)
	r.GET("/home", home.GetHomeHandler)
	r.DELETE("/home/:id", home.DeleteHomeHandler)
	r.Run()
}
