package database

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type DatabaseHandler struct {
	Driver neo4j.DriverWithContext
	Ctx    context.Context
	Config *Neo4jConfiguration
}

var instance *DatabaseHandler
var errConnectivity error

func NewDatabaseHandler() (*DatabaseHandler, error) {
	config := ParseConfiguration()

	// Tentar estabelecer a conexão até 60 segundos
	for i := 0; i < 30; i++ {
		ctx := context.Background()
		driver, err := config.NewDriver()

		if err == nil {
			return &DatabaseHandler{
				Driver: driver,
				Ctx:    ctx,
				Config: config,
			}, nil
		}

		errConnectivity = err // Atualiza o erro de conectividade
		time.Sleep(time.Second)
	}

	return nil, errConnectivity
}

type Neo4jConfiguration struct {
	Url      string
	Username string
	Password string
	Database string
}

func (nc *Neo4jConfiguration) NewDriver() (neo4j.DriverWithContext, error) {
	return neo4j.NewDriverWithContext(nc.Url, neo4j.BasicAuth(nc.Username, nc.Password, ""))
}

func ParseConfiguration() *Neo4jConfiguration {
	database := lookupEnvOrGetDefault("NEO4J_DATABASE", "neo4j")
	if !strings.HasPrefix(lookupEnvOrGetDefault("NEO4J_VERSION", "4"), "4") {
		database = ""
	}
	return &Neo4jConfiguration{
		Url:      lookupEnvOrGetDefault("NEO4J_URI", "bolt://localhost:7687"),
		Username: lookupEnvOrGetDefault("NEO4J_USER", "neo4j"),
		Password: lookupEnvOrGetDefault("NEO4J_PASSWORD", "supersecret"),
		Database: database,
	}
}

func lookupEnvOrGetDefault(key string, defaultValue string) string {
	if env, found := os.LookupEnv(key); !found {
		return defaultValue
	} else {
		return env
	}
}
