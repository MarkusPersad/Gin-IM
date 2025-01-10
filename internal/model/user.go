package model

import (
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

type User struct {
	gorm.Model
	Uuid     string `json:"uuid" gorm:"type:varchar(150);not null;unique_index:idx_uuid;comment:'uuid'"`
	Username string `json:"username" gorm:"unique;not null; comment:'用户名'"`
	Password string `json:"password" gorm:"type:varchar(150);not null; comment:'密码'"`
	Avatar   string `json:"avatar" gorm:"type:varchar(150);comment:'头像'"`
	Email    string `json:"email" gorm:"type:varchar(80);column:email;comment:'邮箱'"`
	Status   int8   `json:"status" gorm:"type:tinyint;default:1;column:status;comment:'状态'"`
	Version  optimisticlock.Version
}
