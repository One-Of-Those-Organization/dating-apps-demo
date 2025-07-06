package table

import (
	"gorm.io/gorm"
)

type Hobby struct {
	gorm.Model
	ID             int       `gorm:"primaryKey"`
	Name           string    `gorm:"column:hobby_name;uniqueIndex"`

	Users          []User    `gorm:"many2many:user_hobbies"`
}
