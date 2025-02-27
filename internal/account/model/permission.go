package model

import "katydid-mp-user/internal/pkg/model"

// BlackList 黑名单
type BlackList struct {
	BlackID    int64  `json:"blackID"`
	AccountID  int64  `json:"accountID"`
	BlackType  int32  `json:"blackType"`  // 黑名单类型
	BlackName  string `json:"blackName"`  // 黑名单名
	BlackState int32  `json:"blackState"` // 黑名单状态
	BlackAt    int64  `json:"blackAt"`    // 黑名单时间
	BlackIP    string `json:"blackIP"`    // 黑名单IP
	History    string `json:"history"`    // 黑名单历史
}

// FreezeList 冻结名单
type FreezeList struct {
	FreezeID int64  `json:"freezeID"`
	ExpireAt int64  `json:"expireAt"` // 解冻时间
	History  string `json:"history"`  // 冻结历史
}

// TODO:GG acc(root) -> perm -> acc(auth) -> user

// TODO:GG acc(perm) -> org -> app -> client <- user <- acc(auth)

// TODO:GG account -> org -> perm -> account<<<<(account[所以account的extra里带有perms])
// TODO:GG account -> org -> app -> client <- account <- user<<<<

// Permission 权限 (Account的) (100*A)
type Permission struct {
	*model.Base

	OrgId int64  `json:"orgId"` // 组织id
	Name  string `json:"name"`  // 权限名

	sub string // 角色
	obj string // 对象
	act string // 操作

	AccountIds []uint64     `json:"accountIds"` // 账号id们
	SubPermIds []Permission `json:"subPermIds"` // 子权限
}
