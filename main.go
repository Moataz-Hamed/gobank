package main

import (
	"log"
)

func main() {
	db, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(db)
	server := NewAPIServer(":3000", db)
	server.Run()
}
