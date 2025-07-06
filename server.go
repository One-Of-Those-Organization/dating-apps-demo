package main

import (
	l "log"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func InitServer(address string, dbFile string, password string) (*Backend, error) {
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
		"./frontend/",
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
		pass:    password,
    }, nil
}

func IsLoggedIn(c *fiber.Ctx, secret string) bool {
    cookie := c.Cookies("jwt")
    if cookie == "" {
        return false
    }
    token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    if err != nil || !token.Valid {
        return false
    }
    return true
}

func InitAPIRoute(backend *Backend) {
	app := backend.app
    api := app.Group("/api")

	cookieJWT := api.Group("/p", jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(backend.pass)},
		TokenLookup: "cookie:jwt",
		ContextKey:  "user",
	}))

	frontendJWT := app.Group("/p", jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(backend.pass)},
		TokenLookup: "cookie:jwt",
		ContextKey:  "user",
		ErrorHandler: func(c *fiber.Ctx, err error) error { return c.Redirect("/login") },
	}))

	app.Static("/style", "./style")

	HandleUserRegister(backend, api)
	HandleUserLogin(backend, api)
	HandleUserEdit(backend, cookieJWT)
	HandleUserInfo(backend, cookieJWT)
	HandleUserLogout(backend, cookieJWT)
	HandleUserStatus(backend, cookieJWT)

	HandleInterestInfoAll(backend, api)

	HandleHobbyInfoAll(backend, api)

	HandleMatchmake(backend, cookieJWT)

	// Frontend routes
	IndexPage(backend, app)
	LoginPage(backend, app)
	RegisterPage(backend, app)
	HomePage(backend, frontendJWT)
	ResultPage(backend, frontendJWT)
}
