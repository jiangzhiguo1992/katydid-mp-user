package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// PostClient handles client creation
func PostClient(c *gin.Context) {
	// TODO: Implement client creation logic
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented",
	})
}
