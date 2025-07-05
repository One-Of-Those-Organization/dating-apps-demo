package table

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID         int       `gorm:"primaryKey"`
	Name       string    `gorm:"column:user_name"`
    FullName   string    `gorm:"column:user_full_name"`
    Password   string    `gorm:"column:user_password" json:"-"`
    Instance   string    `gorm:"column:user_instance"`
    Age        int       `gorm:"column:user_age"`
	Biodata    string    `gorm:"column:user_biodata"` // maybe later if i have time.
    Gender     bool      `gorm:"column:user_gender"`
    Home       string    `gorm:"column:user_home"`

	Hobbies    []Hobby    `gorm:"many2many:HobbyID"`
    Interests  []Interest `gorm:"many2many:InterestID"`
}
