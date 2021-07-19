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
	database := database.NewDb()
	err = database.Connect()
	if err != nil {
		panic(err)
	}

	// Migration
	err = database.DoMigration()
	if err != nil {
		panic(err)
	}
}

func main() {
	initialize()
}
