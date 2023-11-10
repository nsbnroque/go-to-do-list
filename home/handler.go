package home

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/nsbnroque/go-to-do-list/internal/database"
	"github.com/nsbnroque/go-to-do-list/user"
)

func CreateHomeHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if dbHandler == nil || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	userEmail := c.Query("user")

	var homeData Home

	if err := c.ShouldBindJSON(&homeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Erro ao decodificar dados da requisição",
		})
		return
	}

	// Gera um novo UUID para a casa
	homeData.ID = uuid.New()

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (u:User {email: $email})
		MERGE (h:Home {id: $id, name: $name})
		MERGE (u)-[r:LIVES_IN]->(h)
		RETURN u.name as userName, u.email as userEmail, h.name as homeName;`,
		map[string]interface{}{
			"id":    homeData.ID.String(), // Converte UUID para string
			"name":  homeData.Name,
			"email": userEmail,
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
		"message": fmt.Sprintf("Residência criada com sucesso! %v nodes criados em %+v.\n",
			result.Summary.Counters().NodesCreated(),
			result.Summary.ResultAvailableAfter(),
		),
	})
}

func AddResidentToHomeHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if dbHandler == nil || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	userEmail := c.Query("user")
	var newResident user.User
	if err := c.ShouldBindJSON(&newResident); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Erro ao decodificar dados da requisição",
		})
		return
	}

	// Execute a consulta Cypher para associar o usuário à casa e obter os residentes
	result, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (u:User {email: $email})-[:LIVES_IN]->(home:Home)
		MERGE (newResident:User {email: $newResident})
		MERGE (newResident)-[:LIVES_IN]->(home)
		RETURN home, [(home)<-[:LIVES_IN]-(resident:User) | resident] as residents;
		`,
		map[string]interface{}{
			"email":       userEmail,
			"newResident": newResident.Email,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao adicionar usuário",
		})
		return
	}

	var home Home
	for _, record := range result.Records {
		homeNode, ok := record.Values[0].(neo4j.Node)
		if !ok {
			handleInternalError(c, "Erro ao processar resultados da consulta")
			return
		}

		homeIDString, ok := homeNode.Props["id"].(string)
		if !ok {
			handleInternalError(c, "Erro ao processar resultados da consulta")
			return
		}
		_, err := uuid.Parse(homeIDString)
		if err != nil {
			handleInternalError(c, "Erro ao processar resultados da consulta")
			return
		}

		home.Name, ok = homeNode.Props["name"].(string)
		if !ok {
			handleInternalError(c, "Erro ao processar resultados da consulta")
			return
		}

		// Extrair os residentes
		if value, found := record.Get("residents"); found {
			residentNodes := value.([]interface{})
			for _, residentNodeInterface := range residentNodes {
				residentNode, ok := residentNodeInterface.(neo4j.Node)
				if !ok {
					handleInternalError(c, "Erro ao processar resultados da consulta")
					return
				}

				resident := user.User{
					Name:  residentNode.Props["name"].(string),
					Email: residentNode.Props["email"].(string),
				}
				home.Residents = append(home.Residents, resident)
			}
		}
	}

	// Envie uma resposta de sucesso
	c.JSON(http.StatusOK, home)

}

func GetHomeHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if dbHandler == nil || err != nil {
		handleDatabaseError(c, err)
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	// Obter o ID da casa da URL
	id := c.Query("id")

	// Execute a consulta Cypher para associar o usuário à casa e obter os residentes
	result, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (home:Home {id: $id})
		MATCH (resident:User)-[:LIVES_IN]->(home)
		RETURN home, [(home)<-[:LIVES_IN]-(resident:User) | resident] as residents;
		`,
		map[string]interface{}{
			"id": id,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao encontrar casa: %v", err),
		})
		return
	}

	var home Home

	for _, record := range result.Records {
		// Certifique-se de que record.Values[0] é um *neo4j.Node
		homeNode, ok := record.Values[0].(neo4j.Node)
		if !ok {
			handleInternalError(c, "Erro ao processar os dados")
			return
		}

		homeIDString, ok := homeNode.Props["id"].(string)
		if !ok {
			handleInternalError(c, "Erro ao processar os dados")
			return
		}
		home.ID, err = uuid.Parse(homeIDString)
		if err != nil {
			handleInternalError(c, "Erro ao processar os dados")
			return
		}

		home.Name, ok = homeNode.Props["name"].(string)
		if !ok {
			handleInternalError(c, "Erro ao processar os dados")
			return
		}

		// Extrair os residentes
		if value, found := record.Get("residents"); found {
			residentNodes := value.([]interface{})
			for _, residentNodeInterface := range residentNodes {
				residentNode, ok := residentNodeInterface.(neo4j.Node)
				if !ok {
					handleInternalError(c, "Erro ao processar os dados")
					return
				}

				resident := user.User{
					Name:  residentNode.Props["name"].(string),
					Email: residentNode.Props["email"].(string),
				}
				home.Residents = append(home.Residents, resident)
			}
		}
	}

	// Enviar a casa e os residentes como resposta
	c.JSON(http.StatusOK, home)
}

func DeleteHomeHandler(c *gin.Context) {
	dbHandler, err := database.NewDatabaseHandler()

	if dbHandler == nil || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha de conexão com o banco de dados",
		})
		return
	}

	ctx, driver := dbHandler.Ctx, dbHandler.Driver

	// Obter o ID da casa a ser excluída da URL
	id := c.Param("id")

	result, err := neo4j.ExecuteQuery(ctx, driver,
		`MATCH (home:Home {id: $id})
		DETACH DELETE home;
		`,
		map[string]interface{}{
			"id": id,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(dbHandler.Config.Database),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erro ao excluir casa: %v", err),
		})
		return
	}

	// Verifique se alguma casa foi excluída
	if result.Summary.Counters().NodesDeleted() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Casa não encontrada ou já foi excluída",
		})
		return
	}

	// Envie uma resposta de sucesso
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Casa com ID %s excluída com sucesso!", id),
	})
}
