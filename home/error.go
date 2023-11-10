package home

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Função para tratar erro de conexão com o banco de dados
func handleDatabaseError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Falha de conexão com o banco de dados",
	})
}

// Função para tratar erro de requisição inválida
func handleBadRequestError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
	})
}

// Função para tratar erro interno
func handleInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": message,
	})
}
