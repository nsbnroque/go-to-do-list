package home

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/nsbnroque/go-to-do-list/user"
)

func CreateHomeHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userEmail := r.URL.Query().Get("user")

		var homeData Home
		err := json.NewDecoder(r.Body).Decode(&homeData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err), http.StatusBadRequest)
			log.Println(err.Error())
			return
		}

		// Gera um novo UUID para a casa
		homeData.ID = uuid.New()

		// Execute query
		_, err = neo4j.ExecuteQuery(ctx, driver,
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
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao criar a casa: %v", err), http.StatusInternalServerError)
			return
		}

		// Enviar a casa criada como resposta
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(homeData)
	}
}

func AddResidentToHomeHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obter o ID da casa da URL
		userEmail := r.URL.Query().Get("user")

		var resident user.User
		err := json.NewDecoder(r.Body).Decode(&resident)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err), http.StatusBadRequest)
			log.Println(err.Error())
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
				"newResident": resident.Email,
			},
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao adicionar residente à casa: %v", err), http.StatusInternalServerError)
			return
		}

		var home Home

		for _, record := range result.Records {
			// Certifique-se de que record.Values[0] é um *neo4j.Node
			homeNode, ok := record.Values[0].(neo4j.Node)
			if !ok {
				http.Error(w, "Erro ao obter o nó da casa", http.StatusInternalServerError)
				return
			}

			homeIDString, ok := homeNode.Props["id"].(string)
			if !ok {
				http.Error(w, "Erro ao obter o ID da casa", http.StatusInternalServerError)
				return
			}
			home.ID, err = uuid.Parse(homeIDString)
			if err != nil {
				http.Error(w, "Erro ao converter o ID da casa", http.StatusInternalServerError)
				return
			}

			home.Name, ok = homeNode.Props["name"].(string)
			if !ok {
				http.Error(w, "Erro ao obter o nome da casa", http.StatusInternalServerError)
				return
			}

			// Extrair os residentes
			if value, found := record.Get("residents"); found {
				residentNodes := value.([]interface{})
				for _, residentNodeInterface := range residentNodes {
					residentNode, ok := residentNodeInterface.(neo4j.Node)
					if !ok {
						http.Error(w, "Erro ao obter os dados do residente", http.StatusInternalServerError)
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(home)
	}
}

func GetHomeHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obter o ID da casa da URL
		id := r.URL.Query().Get("id")

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
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao encontrar casa: %v", err), http.StatusInternalServerError)
			return
		}

		var home Home

		for _, record := range result.Records {
			// Certifique-se de que record.Values[0] é um *neo4j.Node
			homeNode, ok := record.Values[0].(neo4j.Node)
			if !ok {
				http.Error(w, "Erro ao obter o nó da casa", http.StatusInternalServerError)
				return
			}

			homeIDString, ok := homeNode.Props["id"].(string)
			if !ok {
				http.Error(w, "Erro ao obter o ID da casa", http.StatusInternalServerError)
				return
			}
			home.ID, err = uuid.Parse(homeIDString)
			if err != nil {
				http.Error(w, "Erro ao converter o ID da casa", http.StatusInternalServerError)
				return
			}

			home.Name, ok = homeNode.Props["name"].(string)
			if !ok {
				http.Error(w, "Erro ao obter o nome da casa", http.StatusInternalServerError)
				return
			}

			// Extrair os residentes
			if value, found := record.Get("residents"); found {
				residentNodes := value.([]interface{})
				for _, residentNodeInterface := range residentNodes {
					residentNode, ok := residentNodeInterface.(neo4j.Node)
					if !ok {
						http.Error(w, "Erro ao obter os dados do residente", http.StatusInternalServerError)
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(home)
	}
}
