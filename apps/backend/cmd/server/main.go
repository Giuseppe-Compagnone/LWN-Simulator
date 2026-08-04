package main

import (
	"log"

	"lwn-simulator-backend/internal/server"
	"lwn-simulator-backend/internal/utils"
)

func main() {
	app := server.New()

	privateIP := utils.GetPrivateIP()

	log.Println("Server running:")
	log.Println("  Local:   http://localhost:8080")
	log.Printf("  Network: http://%s:8080\n", privateIP)
	log.Fatal(app.Run(":8080"))
}
