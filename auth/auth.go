package auth

import (
	"context"
	"mlcgo/model"
)

type AuthType int

const (
	OfflineType AuthType = iota
	MicrosoftAuthType
	AuthlibInjectorAuthType
)

type AuthInterface interface {
	Auth(ctx context.Context, loginInfo map[string]string) (*model.UserInfo, error)
	Logout() error
}

// 16,24,32位字符串的话，分别对应AES-128，AES-192，AES-256 加密方法
// key不能泄露
var PwdKey = []byte("a22ga2df1i94asdf")
