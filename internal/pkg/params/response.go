package params

import "github.com/gin-gonic/gin"

type Response struct {
	gin.H
}

func NewResponse(status int, msg string, result any) *Response {
	return &Response{gin.H{"status": status, "msg": msg, "result": result}}
}
