package params

import "github.com/gin-gonic/gin"

func NewResponseJson(c *gin.Context, status int, msg string, result any) {
	c.JSON(status, gin.H{"status": status, "msg": msg, "result": result})
}
