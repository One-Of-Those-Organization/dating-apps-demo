package main
import (
	"github.com/gofiber/fiber/v2"
    "gorm.io/gorm"
)

type Backend struct {
	app       *fiber.App
	db        *gorm.DB
	pass      string
	engine    *DynamicEngine
	address   string
	mode      string
}

