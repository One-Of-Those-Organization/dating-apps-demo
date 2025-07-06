package table

import (
	"gorm.io/gorm"
)

type Interest struct {
	gorm.Model
	ID             int       `gorm:"primaryKey"`
    Name           string    `gorm:"column:interest_name;uniqueIndex"`

	Users          []User    `gorm:"many2many:user_interests"`
}
