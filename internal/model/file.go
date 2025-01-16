package model

import (
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

type File struct {
	gorm.Model
	UploadId   string `json:"uploadId" gorm:"column:uploadid;type:varchar(150); not null; unique;comment:上传ID"`
	Owner      string `json:"owner" gorm:"column:owner;type:varchar(150);not null;comment:owner"`
	Sha1       string `json:"sha1" gorm:"column:sha1;type:varchar(150);not null;uniqueIndex:idx_file;comment:sha1"`
	Md5        string `json:"md5" gorm:"column:md5;type:varchar(150);not null;uniqueIndex:idx_file;comment:md5"`
	ObjectName string `json:"objectName" gorm:"column:objectname;type:varchar(150);not null;comment:objectname"`
	FileName   string `json:"fileName" gorm:"column:filename;type:varchar(150);not null;comment:文件名称"`
	Version    optimisticlock.Version
}
