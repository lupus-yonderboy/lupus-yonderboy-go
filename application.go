package main

import (
	"github.com/lupus-yonderboy/lupus-yonderboy-go/db"
	"github.com/lupus-yonderboy/lupus-yonderboy-go/server"
)

func main() {
	db.Connect()
	server.Start()
}
