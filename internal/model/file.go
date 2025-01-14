package model

import (
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

type File struct {
	gorm.Model
	Owner      string `json:"owner" gorm:"column:owner;type:varchar(150);not null;comment:owner"`
	Sha1       string `json:"sha1" gorm:"column:sha1;type:varchar(150);not null;uniqueIndex:idx_file;comment:sha1"`
	Md5        string `json:"md5" gorm:"column:md5;type:varchar(150);not null;uniqueIndex:idx_file;comment:md5"`
	ObjectName string `json:"objectName" gorm:"column:objectname;type:varchar(150);not null;uniqueIndex:idx_file;comment:objectname"`
	FileType   int8   `json:"fileType" gorm:"column:filetype;type:tinyint;default:0;comment:文件类型"`
	Version    optimisticlock.Version
}
