package main

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidInput 表示用户提交的数据没有通过业务校验。
var ErrInvalidInput = errors.New("invalid signup input")

// SignupInput 表示外部传入的注册数据。
// 输入模型通常保留用户提交时的原始形态，再由业务函数做清洗和校验。
type SignupInput struct {
	Name  string
	Email string
	Age   int
}

// UserProfile 表示系统内部真正要使用的用户资料。
// 它和 SignupInput 分开，是为了避免把外部输入直接当成可信业务数据。
type UserProfile struct {
	DisplayName string
	Email       string
	IsAdult     bool
}

// BuildProfile 校验注册输入，并把它转换成可被系统使用的用户资料。
func BuildProfile(input SignupInput) (UserProfile, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)

	if name == "" {
		return UserProfile{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	if !strings.Contains(email, "@") {
		return UserProfile{}, fmt.Errorf("%w: email must contain @", ErrInvalidInput)
	}

	if input.Age < 13 {
		return UserProfile{}, fmt.Errorf("%w: age must be at least 13", ErrInvalidInput)
	}

	return UserProfile{
		DisplayName: name,
		Email:       strings.ToLower(email),
		IsAdult:     input.Age >= 18,
	}, nil
}
