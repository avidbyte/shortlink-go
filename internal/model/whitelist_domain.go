package model

import "gorm.io/gorm"

type WhitelistDomain struct {
	gorm.Model
	Domain string `gorm:"size:255;uniqueIndex;not null" json:"domain"`
}
