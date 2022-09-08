package main

import "github.com/baiest/sendfileapp/server"

func main() {
	server := server.NewServer("localhost", 3000)
	server.Run()
}
