package model

import (
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

type Group struct {
	gorm.Model
	Uuid      string `json:"uuid" gorm:"column:uuid;type:varchar(150);not null;unique;comment:uuid"`
	UserId    string `json:"userId" gorm:"column:userid;type:varchar(150);not null;unique;comment:群主UUID"`
	GroupName string `json:"groupName" gorm:"column:groupname;type:varchar(32);not null;comment:群名称"`
	Notice    string `json:"notice" gorm:"column:notice;type:varchar(350);comment:群公告"`
	Version   optimisticlock.Version
}
