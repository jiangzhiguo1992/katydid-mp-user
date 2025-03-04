package model

import (
	"katydid-mp-user/internal/pkg/model"
	"katydid-mp-user/pkg/valid"
	"katydid-mp-user/utils"
	"reflect"
	"unicode"
)

// Organization 组织
type Organization struct {
	*model.Base

	OwnAccIds []uint64 `json:"ownAccIds" validate:"required"` // 所属账号们
	ParentIds []uint64 `json:"parentIds"`                     // 父级组织 默认0

	Enable   bool     `json:"enable"`                                  // 是否可用(拉黑等)
	IsPublic bool     `json:"isPublic"`                                // 是否公开
	Kind     uint8    `json:"kind" validate:"required,kind-check"`     // 组织类型
	Become   uint8    `json:"become" validate:"required,become-check"` // 加入方式
	Name     string   `json:"name" validate:"required,name-format"`    // 组织名称
	Display  string   `json:"display" validate:"name-format"`          // 组织显示名称
	Tags     []string `json:"tags"`                                    // 组织标签们

	Children []*Organization `json:"children" gorm:"-:all"` // 子组织列表
	Perms    []*Permission   `json:"perms" gorm:"-:all"`    // 权限列表
	AccIds   []uint64        `json:"accIds" gorm:"-:all"`   // 账号列表
	UserIds  []uint64        `json:"userIds" gorm:"-:all"`  // 成员列表
	AppIds   []uint64        `json:"appIds" gorm:"-:all"`   // 项目列表
}

func NewOrganizationEmpty() *Organization {
	return &Organization{
		Base:      model.NewBaseEmpty(),
		OwnAccIds: make([]uint64, 0),
		ParentIds: make([]uint64, 0),
		Tags:      make([]string, 0),
		Children:  make([]*Organization, 0),
		Perms:     make([]*Permission, 0),
		AccIds:    make([]uint64, 0),
		UserIds:   make([]uint64, 0),
		AppIds:    make([]uint64, 0),
	}
}

func NewOrganizationDefault(
	ownAccIds, parentIds []uint64,
	enable, isPublic bool, kind, become uint8, name, display string, tags []string,
) *Organization {
	return &Organization{
		Base:      model.NewBaseEmpty(),
		OwnAccIds: ownAccIds, ParentIds: parentIds,
		Enable: enable, IsPublic: isPublic, Kind: kind, Become: become, Name: name, Display: display, Tags: tags,
		Children: make([]*Organization, 0),
		Perms:    make([]*Permission, 0),
		AccIds:   make([]uint64, 0),
		UserIds:  make([]uint64, 0),
		AppIds:   make([]uint64, 0),
	}
}

func (o *Organization) ValidFieldRules() valid.FieldValidRules {
	return valid.FieldValidRules{
		valid.SceneAll: valid.FieldValidRule{
			// 组织名称 (1-50)
			"name-format": func(value reflect.Value, param string) bool {
				data := value.String()
				if len(data) < 1 || len(data) > 50 {
					return false
				}
				for _, r := range data {
					if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
						return false
					}
				}
				return true
			},
			"kind-check": func(value reflect.Value, param string) bool {
				data := uint8(value.Uint())
				return data == OrgKindPhysical || data == OrgKindVirtual
			},
			"become-check": func(value reflect.Value, param string) bool {
				data := uint8(value.Uint())
				return data == OrgBecomePublic || data == OrgBecomeApply || data == OrgBecomeInvite
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
				ValidFn: func(value any) bool {
					data, ok := value.(string)
					if !ok {
						return false
					}
					return len(data) <= 1000
				},
			},
			// 简介 (<1000)
			orgExtraKeyDesc: valid.ExtraValidRuleInfo{
				Field: orgExtraKeyDesc,
				ValidFn: func(value any) bool {
					data, ok := value.(string)
					if !ok {
						return false
					}
					return len(data) <= 1000
				},
			},
			// 地址 (<100)*(<1000)
			orgExtraKeyAddresses: valid.ExtraValidRuleInfo{
				Field: orgExtraKeyAddresses,
				ValidFn: func(value any) bool {
					data, ok := value.([]string)
					if !ok {
						return false
					}
					if len(data) > 100 {
						return false
					}
					for _, v := range data {
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
				ValidFn: func(value any) bool {
					data, ok := value.([]string)
					if !ok {
						return false
					}
					if len(data) > 100 {
						return false
					}
					for _, v := range data {
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

func (o *Organization) ValidStructRules(scene valid.Scene, fn valid.FuncReportError) {
	switch scene {
	case valid.SceneAll:
		for _, tag := range o.Tags {
			if len(tag) > 100 {
				fn(o, "Tags", valid.TagFormat, "")
			}
		}
	}
}

func (o *Organization) ValidLocalizeRules() valid.LocalizeValidRules {
	return valid.LocalizeValidRules{
		valid.SceneAll: valid.LocalizeValidRule{
			Rule1: map[valid.Tag]map[valid.FieldName]valid.LocalizeValidRuleParam{
				valid.TagRequired: {
					"OwnAccIds": {"format_s_input_required", false, []any{"own_accounts"}},
					"Name":      {"format_s_input_required", false, []any{"org_name"}},
				},
				valid.TagFormat: {
					"Tags": {"format_tags_err", false, nil},
				},
				"name-format": {
					"Name":    {"format_org_name_err", false, nil},
					"Display": {"format_org_display_err", false, nil},
				},
			}, Rule2: map[valid.Tag]valid.LocalizeValidRuleParam{
				"kind-check":          {"format_org_kind_err", false, nil},
				"become-check":        {"format_org_become_err", false, nil},
				orgExtraKeyWebsiteUrl: {"format_website_err", false, nil},
				orgExtraKeyDesc:       {"format_desc_err", false, nil},
				orgExtraKeyAddresses:  {"format_addresses_err", false, nil},
				orgExtraKeyContacts:   {"format_contacts_err", false, nil},
			},
		},
	}
}

// const
const (
	OrgParentRootId uint64 = 0 // 根组织

	OrgKindPhysical uint8 = 0 // 实体组织 (同时存在数受orgExtraKeyMultiJob影响)
	OrgKindVirtual  uint8 = 1 // 虚拟组织 (能同时存在多个)

	OrgBecomePublic uint8 = 0 // 公开
	OrgBecomeApply  uint8 = 1 // 申请 (只有public有效?)
	OrgBecomeInvite uint8 = 2 // 邀请
)

func (o *Organization) IsTopParent() bool {
	zero := len(o.ParentIds) == 0
	one := len(o.ParentIds) == 1
	return zero || (one && (o.ParentIds[0] == OrgParentRootId))
}

func (o *Organization) IsBotChild() bool {
	return (o.Children == nil) || (len(o.Children) <= 0)
}

// extra
const (
	// TODO:GG 有成员的时候，获取需要各种auth?登录不需要
	orgExtraKeyRootPwd  = "rootPwd"  // 根密码
	orgExtraKeyMultiJob = "multiJob" // 是否允许单用户多任职

	orgExtraKeyWebsiteUrl = "websiteUrl" // 官网
	orgExtraKeyFaviconUrl = "faviconUrl" // 图标
	orgExtraKeyDesc       = "desc"       // 简介
	orgExtraKeyAddresses  = "addresses"  // 地址
	orgExtraKeyContacts   = "contacts"   // 联系方式

	// TODO:GG 支持的Account的认证方式? 支持的Permission的方式?
	// TODO:GG PasswordType, PasswordSalt
)

func (o *Organization) SetRootPwd(pwd *string) {
	o.Extra.SetString(orgExtraKeyRootPwd, pwd)
}

func (o *Organization) GetRootPwd() string {
	data, _ := o.Extra.GetString(orgExtraKeyRootPwd)
	return data
}

func (o *Organization) SetMultiJob(multiJob *bool) {
	o.Extra.SetBool(orgExtraKeyMultiJob, multiJob)
}

func (o *Organization) GetMultiJob() bool {
	data, _ := o.Extra.GetBool(orgExtraKeyMultiJob)
	return data
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
