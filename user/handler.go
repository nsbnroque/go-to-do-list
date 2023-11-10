package user

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/nsbnroque/go-to-do-list/internal/database"
)

func CreateUserHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	var userData User
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Erro ao decodificar dados da requisição",
		})
		return
	}

	hashedPassword, err := HashPassword(userData.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao gerar hash da senha",
		})
		return
	}
	userData.Password = hashedPassword
	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, driver,
		`MERGE (u:User {email: $email})
				SET
					u.name = $name,
					u.password = $password
				RETURN u;
			`,
		map[string]interface{}{
			"name":     userData.Name,
			"password": userData.Password,
			"email":    userData.Email,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao criar usuário",
		})
		return
	}

	// Envie uma resposta de sucesso
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Usuário criado com sucesso! %v nodes criados em %+v.\n",
			result.Summary.Counters().NodesCreated(),
			result.Summary.ResultAvailableAfter(),
		),
	})

}

func UpdateUserHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	var userData User
	if err := c.ShouldBindJSON(&userData); err != nil {
		log.Println("Error: Failed to bind JSON", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	hashedPassword, _ := HashPassword(userData.Password)
	userData.Password = hashedPassword

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (u:User {email: $email})
			SET
				u.name = $name,
				u.password = $password
			RETURN u;
		`,
		map[string]interface{}{
			"name":     userData.Name,
			"password": userData.Password,
			"email":    userData.Email,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao atualizar usuário",
		})
		return
	}

	if len(result.Records) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Usuário não encontrado",
		})
		return
	}

	// Envie uma resposta de sucesso
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Usuário atualizado com sucesso! %v nodes criados em %+v.\n",
			result.Summary.Counters().NodesCreated(),
			result.Summary.ResultAvailableAfter(),
		),
	})
}

func FindAllUsersHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if dbHandler == nil || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, driver,
		"MATCH (u:User) RETURN u.name AS name, u.email AS email, u.password AS password",
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao executar a consulta: %v", err),
		})
		return
	}

	// Processar os resultados
	var users []User

	for _, record := range result.Records {
		name, _ := record.Get("name")
		email, _ := record.Get("email")
		password, _ := record.Get("password")

		user := User{
			Name:     name.(string),
			Email:    email.(string),
			Password: password.(string),
		}

		users = append(users, user)
	}

	// Enviar a lista de usuários como resposta
	c.JSON(http.StatusOK, users)
}

func FindByEmailHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	email := c.Query("email")
	result, err := neo4j.ExecuteQuery(ctx, driver,
		"MATCH (u:User{email: $email}) RETURN u.name AS name, u.email AS email",
		map[string]interface{}{"email": email},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao executar a consulta: %v", err),
		})
		return
	}

	for _, record := range result.Records {
		name, _ := record.Get("name")
		email, _ := record.Get("email")

		user := User{
			Name:  name.(string),
			Email: email.(string),
		}

		// Enviar a lista de usuários como resposta
		c.JSON(http.StatusOK, user)
	}
}

func DeleteByEmailHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	email := c.Query("email")
	_, err = neo4j.ExecuteQuery(ctx, driver,
		`MATCH (u:User{email: $email}) 
		DETACH DELETE u`,
		map[string]interface{}{"email": email},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao excluir usuário: %v", err),
		})
		return
	}

	// Envie uma resposta de sucesso
	c.JSON(http.StatusOK, gin.H{
		"message": "Usuário excluído com sucesso!",
	})
}

func UpdateScore(ctx context.Context, driver neo4j.DriverWithContext, database string, email, task string) error {
	query := `
		MATCH (u:User {email: $email})-[:HAS_HOME]->(h:Home)-[:HAS_TASK]->(t:Task {nome: $task})
		ON MATCH SET u.score = u.score + t.reward
		RETURN t.reward AS reward, u AS user
	`

	_, err := neo4j.ExecuteQuery(ctx, driver, query,
		map[string]interface{}{"email": email, "task": task},
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(database),
	)

	if err != nil {
		return fmt.Errorf("Erro ao atualizar pontuação: %v", err)
	}

	//	for _, record := range result.Records {
	//	user, _ := record.Get("user")

	// Lógica para processar o usuário, como enviar para um canal ou realizar outra ação
	// ...

	//}

	return nil
}
