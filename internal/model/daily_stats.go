package model

type DailyStat struct {
	BaseModel
	ShortLinkID uint   `gorm:"index"`
	Date        string `gorm:"type:date;index"` // YYYY-MM-DD
	PV          int64  `gorm:"default:0"`
	UV          int64  `gorm:"default:0"`
}
