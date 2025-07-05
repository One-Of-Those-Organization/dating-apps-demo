package main

import (
	l "log"
)

func main() {
	const dbFile = "db/databse.sqlite"
	const address = "0.0.0.0:3000"
	const password = "password"
	app, err := InitServer(address, dbFile, password)
	if err != nil {
		l.Printf("Failed to Init the server, %v.", err)
		return
	}
	InitAPIRoute(app)
	if err := app.app.Listen(app.address); err != nil {
		l.Fatal("Server failed to start: ", err)
	}
}
