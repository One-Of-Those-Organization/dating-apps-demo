package main

import (
	"github.com/gofiber/fiber/v2"
)

func IndexPage(bend *Backend, route fiber.Router) {
	route.Get("/", func (c *fiber.Ctx) error {
		logged := IsLoggedIn(c, bend.pass)

		bend.engine.ClearCache()
		return c.Render("index", fiber.Map{
			"LoggedIn": logged,
		})
	})
}

func LoginPage(bend *Backend, route fiber.Router) {
	route.Get("/login", func (c *fiber.Ctx) error {
		logged := IsLoggedIn(c, bend.pass)
		if logged {
			return c.Redirect("/p/home")
		}

		bend.engine.ClearCache()
		return c.Render("login", fiber.Map{
			"LoggedIn": logged,
		})
	})
}

func RegisterPage(bend *Backend, route fiber.Router) {
	route.Get("/register", func (c *fiber.Ctx) error {
		logged := IsLoggedIn(c, bend.pass)
		if logged {
			return c.Redirect("/p/home")
		}

		bend.engine.ClearCache()
		return c.Render("register", fiber.Map{
			"LoggedIn": logged,
		})
	})
}

func HomePage(bend *Backend, route fiber.Router) {
	route.Get("/home", func (c *fiber.Ctx) error {
		_, err := GetJWT(c)
		if err != nil {
			return c.Redirect("/login")
		}
		
		bend.engine.ClearCache()
		return c.Render("home", fiber.Map{})
	})
}

func ResultPage(bend *Backend, route fiber.Router) {
	route.Get("/result", func (c *fiber.Ctx) error {
		_, err := GetJWT(c)
		if err != nil {
			return c.Redirect("/login")
		}

		bend.engine.ClearCache()
		return c.Render("result", fiber.Map{
			"Title": "Result Page",
		})
	})
}
