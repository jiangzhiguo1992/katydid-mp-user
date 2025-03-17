package model

import "katydid-mp-user/internal/pkg/model"

type Token struct {
	*model.Base
	AccountID uint64  `json:"accountId"` // 账号ID
	OwnKind   OwnKind `json:"ownKind"`   // 账号类型
	OwnID     uint64  `json:"ownId"`     // 账号ID
	DeviceID  string  `json:"deviceId"`  // 设备ID
	Token     string  `json:"token"`     // token

	ExpireAt int64 `json:"expireAt"` // 过期时间
}

func NewTokenEmpty() *Token {
	return &Token{
		Base: model.NewBase(make(map[string]any)),
	}
}

func NewToken(
	accountID uint64, ownKind OwnKind, ownID uint64,
	deviceID string, token string, expireAt int64,
) *Token {
	return &Token{
		Base:      model.NewBase(make(map[string]any)),
		AccountID: accountID, OwnKind: ownKind, OwnID: ownID,
		DeviceID: deviceID, Token: token, ExpireAt: expireAt,
	}
}
