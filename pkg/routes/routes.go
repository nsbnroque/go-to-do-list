package routes

import (
	"context"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nsbnroque/go-to-do-list/home"
	"github.com/nsbnroque/go-to-do-list/internal/database"
	"github.com/nsbnroque/go-to-do-list/task"
	"github.com/nsbnroque/go-to-do-list/user"
)

func HandleRequests() {
	ctx := context.Background()
	configuration := database.ParseConfiguration()
	driver, err := configuration.NewDriver()
	if err != nil {
		log.Fatal(err)
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		log.Fatalf("Falha na conexão: %v", err)
	} else {
		log.Println("Conexão bem-sucedida!")
	}

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PATCH", "DELETE"}
	r.Use(cors.New(config))

	r.POST("/users", func(c *gin.Context) {
		user.CreateUserHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.GET("/users", func(c *gin.Context) {
		user.FindAllUsersHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.GET("/users/find", func(c *gin.Context) {
		user.FindByEmailHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.PUT("/users", func(c *gin.Context) {
		user.UpdateUserHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.POST("/tasks", func(c *gin.Context) {
		task.CreateTaskHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.PUT("/tasks", func(c *gin.Context) {
		task.ChangeTaskHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.DELETE("/tasks", func(c *gin.Context) {
		task.DeleteTaskHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.GET("/tasks", func(c *gin.Context) {
		task.GetTasksForUserHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.POST("/homes", func(c *gin.Context) {
		home.CreateHomeHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.PATCH("/homes", func(c *gin.Context) {
		home.AddResidentToHomeHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.GET("/homes", func(c *gin.Context) {
		home.GetHomeHandler(ctx, driver, database.ParseConfiguration().Database)(c.Writer, c.Request)
	})
	r.Run()
}
