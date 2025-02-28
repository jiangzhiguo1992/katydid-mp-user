package params

import (
	"github.com/gin-gonic/gin"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/valid"
)

func RequestBind(c *gin.Context, obj any) *err.CodeErrs {
	e := c.Bind(obj)
	if e != nil {
		return err.Match(e)
	}
	cErr := valid.Check(obj, valid.SceneBind)
	return cErr
}
