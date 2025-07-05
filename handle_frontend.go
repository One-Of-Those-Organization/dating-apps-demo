package main

import (
	"github.com/gofiber/fiber/v2"
)

func HandleHello(bend *Backend, route fiber.Router) {
	route.Get("halo", func (c *fiber.Ctx) error {
		bend.engine.ClearCache()
		return c.Render("index", fiber.Map{
			"Halo": "Dunia",
		})

	})
}
