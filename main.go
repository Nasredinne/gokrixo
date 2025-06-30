package main

import (
	"fmt"
	"log"
)

func main() {

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer("0.0.0.0:3000", store)
	fmt.Println("Coonectect to server ")
	server.Run()
}
