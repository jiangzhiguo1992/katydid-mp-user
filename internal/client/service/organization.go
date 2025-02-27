package service

import (
	"katydid-mp-user/internal/client/model"
	"katydid-mp-user/internal/pkg/service"
	"katydid-mp-user/pkg/err"
)

type Team struct {
	service.Base
}

func NewServiceTeam() *Team {
	return &Team{
		service.Base{
			//ctx: service.NewCtx(0, 0, nil),
		},
	}
}

//func (t *Organization) Add() {
//
//}

func (t *Team) AddTeam(team *model.Organization) *err.CodeErrs {
	// 检查parent
	//if team.OwnAccIds == model.OrgParentRootId {
	//	return err.MatchByMsg("所属团队无效")
	//}
	//parent, errs := GetTeam(team.ParentId)
	//if errs != nil {
	//	return errs
	//} else if parent == nil {
	//	return err.MatchByMsg("父节点不存在")
	//}
	// TODO:GG DB insert
	return nil
}

func GetTeam(id uint64) (*model.Organization, *err.CodeErrs) {
	// TODO:GG cache
	// TODO:GG DB query
	team := model.NewOrganizationEmpty()
	team.Id = id
	return team, nil
}

//import (
//	"katydid_base_api/internal/client/model"
//	"katydid_base_api/internal/client/repository/db"
//	"katydid_base_api/internal/pkg/utils"
//	"katydid_base_api/tools"
//)
//
//func AddTeam(instance *model.Organization) *tools.CodeError {
//	// TODO:GG permission
//
//	if instance.ParentId != model.OrgParentRootId {
//		var hide *bool = nil
//		// TODO:GG 删除的不能添加，除非是管理员？
//		parent, codeError := db.SelectTeam(instance.ParentId, hide) // TODO:GG ?
//		if parent == nil {
//			return utils.MatchErrorByCode(utils.ErrorCodeDBForeignNoFind)
//		} else if codeError != nil {
//			return codeError
//		}
//	}
//	return db.InsertTeam(instance)
//}
//
//func DelTeam(id uint64) (bool, *tools.CodeError) {
//	// TODO:GG permission
//	// TODO:GG cache
//	return db.DeleteTeam(id)
//}
//
//func HideTeam(id uint64, by int64) (bool, *tools.CodeError) {
//	// TODO:GG permission
//	// TODO:GG cache
//	//client, err := db.QueClient(id)
//	//if client == nil {
//	//	return false, err
//	//}
//	//if client.DeleteBy != 0 {
//	//	if (client.DeleteBy > 0) && (by > 0) {
//	//		// TODO:GG 重复操作
//	//	} else if (client.DeleteBy < 0) && (by < 0) {
//	//		// TODO:GG 重复操作
//	//	}
//	//	//return false, utils.MatchErrorByCode(tools.ErrorCodeDBDelBy)
//	//}
//	return db.HideTeam(id, 0)
//}
//
//func UpdTeamParentId(team *model.Organization, parentId uint64) (bool, *tools.CodeError) {
//	// TODO:GG permission
//	// TODO:GG cache
//	return db.UpdateTeamParentId(team, parentId)
//}
//
//func UpdTeamEnable(team *model.Organization, enable bool) (bool, *tools.CodeError) {
//	// TODO:GG permission
//	// TODO:GG cache
//	return db.UpdateTeamEnable(team, enable)
//}
//
//func UpdTeamName(team *model.Organization, name string) (bool, *tools.CodeError) {
//	// TODO:GG permission
//	// TODO:GG cache
//	return db.UpdateTeamName(team, name)
//}
//
//func GetTeam(id uint64) (*model.Organization, *tools.CodeError) {
//	// TODO:GG permission
//	// TODO:GG cache
//	var hide *bool = nil
//	team, codeError := db.SelectTeam(id, hide)
//	return team, codeError
//}
//
//func GetTeamList(parentId *uint64, enable *bool, name *string, hide *bool) ([]*model.Application, *tools.CodeError) {
//	// TODO:GG permission
//	// TODO:GG cache
//	teams, codeError := db.SelectTeams(parentId, enable, name, hide)
//	return teams, codeError
//}
