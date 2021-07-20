package main

import (
	"github.com/joho/godotenv"
	"github.com/sisu-network/deyes/database"
)

func initialize() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Connect DB
	db := database.NewDb()
	err = db.Connect()
	if err != nil {
		panic(err)
	}

	// Migration
	err = db.DoMigration()
	if err != nil {
		panic(err)
	}

	// testDb(db)
}

func main() {
	initialize()
}
