package main

import (
	l "log"

	// "github.com/gofiber/fiber/v2"
	// "gorm.io/gorm"
	// "gorm.io/driver/sqlite"
)

func main() {
	const dbFile = "databse.sqlite"
	db, err := ReadDB(dbFile)
	if err != nil {
		l.Printf("Failed to opent the db, %v.", err)
		return
	}
	err = MigrateDB(db)
	if err != nil {
		l.Printf("Failed to migrate the db, %v.", err)
		return
	}
}
