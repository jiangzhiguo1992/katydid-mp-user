package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/valid"
	"katydid-mp-user/utils"
	"reflect"
	"unicode"
)

const (
	OrganizationParentRoot uint64 = 0 // 根组织

	// TODO:GG 有成员的时候，获取需要各种auth?登录不需要
	orgExtraKeyRootPwd = "rootPwd" // 根密码

	orgExtraKeyWebsiteUrl = "website"   // 官网
	orgExtraKeyFaviconUrl = "favicon"   // 图标
	orgExtraKeyDesc       = "desc"      // 简介
	orgExtraKeyAddresses  = "addresses" // 地址
	orgExtraKeyContacts   = "contacts"  // 联系方式

	// TODO:GG 支持的Account的认证方式? 支持的Permission的方式?
	// TODO:GG PasswordType, PasswordSalt
)

// Organization 组织
type Organization struct {
	*model.Base

	ParentId uint64   `json:"parentId"`                     // 父级组织 默认0
	OwnerIds []uint64 `json:"ownerIds" validate:"required"` // 所属用户们

	Enable   bool     `json:"enable"`                               // 是否可用
	IsPublic bool     `json:"isPublic"`                             // 是否公开
	Name     string   `json:"name" validate:"required,name-format"` // 组织名称
	Display  string   `json:"display" validate:"display-format"`    // 组织显示名称
	Tags     []string `json:"tags"`                                 // 组织标签们

	Children []*Organization `json:"children" gorm:"-:all"` // 子组织列表
	Apps     []*Application  `json:"apps" gorm:"-:all"`     // 项目列表
	// TODO:GG Permission + Account + User
}

func NewOrganizationEmpty() *Organization {
	return &Organization{
		Base:     model.NewBaseEmpty(),
		ParentId: OrganizationParentRoot,
		Children: []*Organization{},
		Apps:     []*Application{},
	}
}

func NewOrganizationDefault(
	parentId uint64, enable bool, name string,
) *Organization {
	return &Organization{
		Base:     model.NewBaseEmpty(),
		ParentId: parentId, Enable: enable, Name: name,
		Children: []*Organization{},
		Apps:     []*Application{},
	}
}

func (o *Organization) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 组织名称 (1-50)
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

func (o *Organization) ValidExtraRules() (utils.KSMap, valid.ExtraValidRules) {
	return o.Extra, valid.ExtraValidRules{
		valid.SceneAll: valid.ExtraValidRule{
			// 官网 (<1000)
			orgExtraKeyWebsiteUrl: valid.ExtraValidRuleInfo{
				Field: orgExtraKeyWebsiteUrl,
				ValidFn: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 简介 (<1000)
			orgExtraKeyDesc: valid.ExtraValidRuleInfo{
				Field: orgExtraKeyDesc,
				ValidFn: func(value interface{}) bool {
					v, ok := value.(string)
					if !ok {
						return false
					}
					return len(v) <= 1000
				},
			},
			// 地址 (<100)*(<1000)
			orgExtraKeyAddresses: valid.ExtraValidRuleInfo{
				Field: orgExtraKeyAddresses,
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
			orgExtraKeyContacts: valid.ExtraValidRuleInfo{
				Field: orgExtraKeyContacts,
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

func (o *Organization) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"Name": {"format_s_input_required", false, []any{"org_name"}},
					//"ParentId": {"format_s_input_required", false, []any{"org_parent"}},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"name-format":         {"format_org_name_err", false, nil},
				orgExtraKeyWebsiteUrl: {"format_website_err", false, nil},
				orgExtraKeyDesc:       {"format_desc_err", false, nil},
				orgExtraKeyAddresses:  {"format_addresses_err", false, nil},
				orgExtraKeyContacts:   {"format_contacts_err", false, nil},
			},
		},
	}
}

func (o *Organization) SetWebsiteUrl(website *string) {
	o.Extra.SetString(orgExtraKeyWebsiteUrl, website)
}

func (o *Organization) GetWebsiteUrl() string {
	data, _ := o.Extra.GetString(orgExtraKeyWebsiteUrl)
	return data
}

func (o *Organization) SetFaviconUrl(website *string) {
	o.Extra.SetString(orgExtraKeyFaviconUrl, website)
}

func (o *Organization) GetFaviconUrl() string {
	data, _ := o.Extra.GetString(orgExtraKeyFaviconUrl)
	return data
}

func (o *Organization) SetDesc(desc *string) {
	o.Extra.SetString(orgExtraKeyDesc, desc)
}

func (o *Organization) GetDesc() string {
	data, _ := o.Extra.GetString(orgExtraKeyDesc)
	return data
}

func (o *Organization) SetAddresses(addresses *[]string) {
	o.Extra.SetStringSlice(orgExtraKeyAddresses, addresses)
}

func (o *Organization) GetAddresses() []string {
	data, _ := o.Extra.GetStringSlice(orgExtraKeyAddresses)
	return data
}

func (o *Organization) SetContacts(contacts *[]string) {
	o.Extra.SetStringSlice(orgExtraKeyContacts, contacts)
}

func (o *Organization) GetContacts() []string {
	data, _ := o.Extra.GetStringSlice(orgExtraKeyContacts)
	return data
}

func (o *Organization) IsTopParent() bool {
	return o.ParentId == OrganizationParentRoot
}

func (o *Organization) IsBotChild() bool {
	return (o.Children == nil) || (len(o.Children) <= 0)
}

func (o *Organization) GetAllApps() map[uint64][]*Application {
	apps := map[uint64][]*Application{}
	apps[o.Id] = o.Apps
	for _, child := range o.Children {
		childApps := child.GetAllApps()
		for k, v := range childApps {
			apps[k] = v
		}
	}
	return apps
}
