package main

import (
	l "log"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func InitServer(address string, dbFile string) (*Backend, error) {
	db, err := ReadDB(dbFile)
	if err != nil {
		l.Printf("Failed to opent the db, %v.\n", err)
		return nil, err
	}
	err = MigrateDB(db)
	if err != nil {
		l.Printf("Failed to migrate the db, %v.\n", err)
		return nil, err
	}

    engine := NewDynamicEngine([]string{
        "./static/",
	}, ".html")
    app := fiber.New(fiber.Config{
        AppName: "Dating apps demo.",
        Views: engine,
    })

    return &Backend{
        app:     app,
        db:      db,
        engine:  engine,
        address: address,
        mode:    "http",
		pass:    "password",
    }, nil
}

func InitAPIRoute(backend *Backend) {
	app := backend.app
    api := app.Group("/api")

    protected := api.Group("/p", jwtware.New(jwtware.Config{
        SigningKey: jwtware.SigningKey{Key: []byte(backend.pass)},
    }))
	cookieJWT := api.Group("/c", jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(backend.pass)},
		TokenLookup: "cookie:jwt",
		ContextKey:  "user",
	}))
    app.Static("/static", "./static")

	HandleUserRegister(backend, api)
}
