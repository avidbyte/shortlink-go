package model

type DailyStat struct {
	BaseModel
	ShortLinkID uint   `gorm:"index:uniq_shortlinkid_date,unique"`           // part 1 of unique index
	Date        string `gorm:"type:date;index:uniq_shortlinkid_date,unique"` // part 2
	PV          uint64 `gorm:"default:0"`
	UV          uint64 `gorm:"default:0"`
}
