package service

import "errors"

var (
	ErrUsernameTaken      = errors.New("该用户名已被占用")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUserNotFound       = errors.New("用户不存在")
)
