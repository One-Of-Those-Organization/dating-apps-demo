package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "dating-apps/table"
)

func ReadDB(dbFile string) (*gorm.DB, error){
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return db, nil
}

func MigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(&table.User{})
	if err != nil { return err }
	err = db.AutoMigrate(&table.Hobby{})
	if err != nil { return err }
	err = db.AutoMigrate(&table.Interest{})
	if err != nil { return err }
	return nil
}
