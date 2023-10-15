package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func CreateUserHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var userData User
		err := json.NewDecoder(r.Body).Decode(&userData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err), http.StatusBadRequest)
			log.Println(err.Error())
			return
		}

		hashedPassword, _ := HashPassword(userData.Password)

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
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao criar usuário: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Usuário criado com sucesso! %v nodes criados em %+v.\n",
			result.Summary.Counters().NodesCreated(),
			result.Summary.ResultAvailableAfter(),
		)
	}
}

func UpdateUserHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var userData User
		err := json.NewDecoder(r.Body).Decode(&userData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao decodificar dados da requisição: %v", err), http.StatusBadRequest)
			log.Println(err.Error())
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
			neo4j.ExecuteQueryWithDatabase(database),
		)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao atualizar usuário: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Usuário atualizado com sucesso! %v nodes criados em %+v.\n",
			result.Summary.Counters().NodesCreated(),
			result.Summary.ResultAvailableAfter(),
		)
	}
}

func FindAllUsersHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Execute query
		result, err := neo4j.ExecuteQuery(ctx, driver,
			"MATCH (u:User) RETURN u.name AS name, u.email AS email, u.password AS password",
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao executar a consulta: %v", err), http.StatusInternalServerError)
			log.Println(err.Error())
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func FindByEmailHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		result, err := neo4j.ExecuteQuery(ctx, driver,
			"MATCH (u:User{email: $email}) RETURN u.name AS name, u.email AS email",
			map[string]interface{}{"email": email},
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao executar a consulta: %v", err), http.StatusInternalServerError)
			log.Println(err.Error())
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
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)

		}

	}
}

func DeleteByEmailHandler(ctx context.Context, driver neo4j.DriverWithContext, database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		_, err := neo4j.ExecuteQuery(ctx, driver,
			`MATCH (u:User{email: $email}) 
			DETACH DELETE u`,
			map[string]interface{}{"email": email},
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase(database),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao excluir usuário: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
