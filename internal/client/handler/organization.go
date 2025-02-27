package handler

import (
	"github.com/gin-gonic/gin"
	"katydid-mp-user/internal/client/model"
	"katydid-mp-user/internal/client/service"
	"katydid-mp-user/internal/pkg/handler"
	"katydid-mp-user/internal/pkg/params"
	"net/http"
)

type Team struct {
	*handler.Base
}

func NewTeam() *Team {
	return &Team{
		Base: handler.NewBase(),
	}
}

func (th *Team) PostTeam(c *gin.Context) {
	team := model.NewOrganizationEmpty()
	errs := params.RequestBind(c, team)
	if errs != nil {
		params.ResponseErr(c, http.StatusBadRequest, errs)
		return
	}
	// TODO:GG 账号权限?  一般是root才有权限Add (也就是蒋)
	ts := &service.Team{}
	errs = ts.AddTeam(team)
	if errs != nil {
		params.ResponseErr(c, http.StatusBadRequest, errs)
		return
	}
	//codeError := service.AddClient(client)
	//if codeError != nil {
	//	c.String(http.StatusInternalServerError, codeError.Error())
	//	return
	//}

	params.Response(c, http.StatusOK, "", team)
}

//func PostTeam(c *gin.Context) {
//	instance := model.NewOrganizationEmpty()
//	err := c.BindJSON(instance)
//	if err != nil {
//		c.String(http.StatusBadRequest, "Invalid request")
//		return
//	}
//	codeError := service.AddTeam(instance)
//	if codeError != nil {
//		c.String(http.StatusInternalServerError, codeError.Error())
//		return
//	}
//	c.JSON(http.StatusOK, instance)
//}
//
//// GetTeam godoc
//// @Summary      Show an account
//// @Description  get string by ID
//// @Tags         team
//// @Accept       json
//// @Produce      json
//// @Param        id   path      int  true  "Account ID"
//// @Success      200  {object}  model.Organization
//// @Failure      400  {string}  1
//// @Failure      404  {string}  2
//// @Failure      500  {string}  3
//// @Router       /team/{id} [get]
//func GetTeam(c *gin.Context) {
//	pId := c.Param("id")
//	id, err := strconv.ParseUint(pId, 10, 64)
//	if err != nil {
//		c.String(http.StatusBadRequest, "Invalid ID")
//		return
//	}
//	instance, codeError := service.GetTeam(id)
//	if codeError != nil {
//		c.JSON(http.StatusNotFound, instance)
//	}
//	c.JSON(http.StatusOK, instance)
//}
