package model

import "time"

type Software struct {
	//gorm:"size:100;not null"：管数据库映射。
	//GORM 根据这个 tag 决定：映射到数据库时，这一列的类型是 VARCHAR(100)，并且建表时带 NOT NULL 约束。
	ID          uint      `json:"id" gorm:"size:12;primary_key"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:255;not null;type:text"`
	Version     string    `json:"version" gorm:"size:50;not null"`

	Versions 	[]SoftwareVersion `json:"versions" gorm:"foreignkey:SoftwareID"`//外键关联

	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Software) TableName() string {
	return "software"
}
