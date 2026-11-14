package model

import "time"

type SoftwareVersion struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	SoftwareID  uint      `json:"software_id" gorm:"not null;uniqueIndex:idx_software_version"`
	Software    Software  `json:"-" gorm:"foreignkey:SoftwareID"`
	VersionID   uint      `json:"version_id" gorm:"not null;uniqueIndex:idx_software_version"`
	Version     string    `json:"version" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"type:text"`
	DownloadURL string    `json:"download_url" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SoftwareVersion) TableName() string {
	return "software_version"
}
