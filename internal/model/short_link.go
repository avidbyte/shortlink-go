package model

type ShortLink struct {
	BaseModel
	ShortCode    string `gorm:"uniqueIndex;size:32;not null" json:"shortCode"`
	TargetURL    string `gorm:"size:2048;not null" json:"targetUrl"`
	RedirectCode int    `gorm:"default:302" json:"redirectCode"`
	Disabled     bool   `json:"disabled" json:"disabled"`
	TotalPV      int64  `gorm:"default:0" json:"totalPv"`
	TotalUV      int64  `gorm:"default:0" json:"totalUv"`
	UvHLLBackup  []byte `gorm:"type:blob" json:"-"`
}
