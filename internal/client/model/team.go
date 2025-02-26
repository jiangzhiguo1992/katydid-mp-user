package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/valid"
	"katydid-mp-user/utils"
	"reflect"
	"unicode"
)

const (
	TeamParentRoot uint64 = 0 // 根团队

	teamExtraKeyWebsite   = "website"   // 官网
	teamExtraKeyDesc      = "desc"      // 简介
	teamExtraKeyAddresses = "addresses" // 地址
	teamExtraKeyContacts  = "contacts"  // 联系方式
)

type Team struct {
	*model.Base

	ParentId uint64 `json:"parentId" validate:"required"` // 父级团队 默认0

	Enable bool   `json:"enable"`                               // 是否可用
	Name   string `json:"name" validate:"required,name-format"` // 团队名称

	Children []*Team   `json:"children" gorm:"-:all"` // 子团队列表
	Clients  []*Client `json:"clients" gorm:"-:all"`  // 项目列表
}

func NewTeamEmpty() *Team {
	return &Team{Base: model.NewBaseEmpty()}
}

func NewTeamDefault(
	parentId uint64, enable bool, name string,
) *Team {
	return &Team{
		Base:     model.NewBaseEmpty(),
		ParentId: parentId, Enable: enable, Name: name,
		Children: []*Team{},
		Clients:  []*Client{},
	}
}

func (t *Team) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 团队名称 (1-50)
			"name-format": func(value reflect.Value, param string) bool {
				name := value.String()
				if len(name) < 1 || len(name) > 50 {
					return false
				}
				for _, r := range name {
					if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
						return false
					}
				}
				return true
			},
		},
	}
}

func (t *Team) ValidExtraRules() (utils.KSMap, valid.ExtraValidRules) {
	return t.Extra, valid.ExtraValidRules{
		valid.SceneAll: valid.ExtraValidRule{
			// 官网 (<1000)
			teamExtraKeyWebsite: valid.ExtraValidRuleInfo{
				Field: teamExtraKeyWebsite,
				ValidFn: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 简介 (<1000)
			teamExtraKeyDesc: valid.ExtraValidRuleInfo{
				Field: teamExtraKeyDesc,
				ValidFn: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 地址 (<100)*(<1000)
			teamExtraKeyAddresses: valid.ExtraValidRuleInfo{
				Field: teamExtraKeyAddresses,
				ValidFn: func(value interface{}) bool {
					vs, ok := value.([]string)
					if !ok {
						return false
					}
					if len(vs) > 100 {
						return false
					}
					for _, v := range vs {
						if len(v) > 1000 {
							return false
						}
					}
					return true
				},
			},
			// 联系方式 (<100)*(<1000)
			teamExtraKeyContacts: valid.ExtraValidRuleInfo{
				Field: teamExtraKeyContacts,
				ValidFn: func(value interface{}) bool {
					vs, ok := value.([]string)
					if !ok {
						return false
					}
					if len(vs) > 100 {
						return false
					}
					for _, v := range vs {
						if len(v) > 1000 {
							return false
						}
					}
					return true
				},
			},
		},
	}
}

func (t *Team) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"Name": {"format_s_input_required", false, []any{"team_name"}},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"name-format":         {"format_team_name_err", false, nil},
				teamExtraKeyWebsite:   {"format_website_err", false, nil},
				teamExtraKeyDesc:      {"format_desc_err", false, nil},
				teamExtraKeyAddresses: {"format_addresses_err", false, nil},
				teamExtraKeyContacts:  {"format_contacts_err", false, nil},
			},
		},
	}
}

func (t *Team) SetWebsite(website *string) {
	t.Extra.SetString(teamExtraKeyWebsite, website)
}

func (t *Team) GetWebsite() string {
	data, _ := t.Extra.GetString(teamExtraKeyWebsite)
	return data
}

func (t *Team) SetDesc(desc *string) {
	t.Extra.SetString(teamExtraKeyDesc, desc)
}

func (t *Team) GetDesc() string {
	data, _ := t.Extra.GetString(teamExtraKeyDesc)
	return data
}

func (t *Team) SetAddresses(addresses *[]string) {
	t.Extra.SetStringSlice(teamExtraKeyAddresses, addresses)
}

func (t *Team) GetAddresses() []string {
	data, _ := t.Extra.GetStringSlice(teamExtraKeyAddresses)
	return data
}

func (t *Team) SetContacts(contacts *[]string) {
	t.Extra.SetStringSlice(teamExtraKeyContacts, contacts)
}

func (t *Team) GetContacts() []string {
	data, _ := t.Extra.GetStringSlice(teamExtraKeyContacts)
	return data
}

func (t *Team) IsTopParent() bool {
	return t.ParentId == TeamParentRoot
}

func (t *Team) IsBotChild() bool {
	return (t.Children == nil) || (len(t.Children) <= 0)
}

func (t *Team) GetAllClients() map[uint64][]*Client {
	clients := map[uint64][]*Client{}
	clients[t.Id] = t.Clients
	for _, child := range t.Children {
		childClients := child.GetAllClients()
		for k, v := range childClients {
			clients[k] = v
		}
	}
	return clients
}
