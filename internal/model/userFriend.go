package model

import (
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

type UserFriend struct {
	gorm.Model
	UserId   string `json:"userId" gorm:"column:userid;type:varchar(150);not null;uniqueIndex:idx_uuid;comment:用户ID"`
	FriendId string `json:"friendId" gorm:"column:friendid;type:varchar(150);not null;uniqueIndex:idx_uuid;comment:好友ID"`
	Version  optimisticlock.Version
}
