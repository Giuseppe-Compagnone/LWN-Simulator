package main

import (
	"flag"
	"fmt"
	"log"

	"lwn-simulator-backend/internal/server"
	"lwn-simulator-backend/internal/utils"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "8080", "Port to start the server on")
	flag.StringVar(&port, "p", "8080", "Port to start the server on (short)")
	flag.Parse()

	app := server.New(port)

	privateIP := utils.GetPrivateIP()

	address := fmt.Sprintf(":%s", port)

	log.Println("Server running:")
	log.Printf("  Local:   http://localhost:%s\n", port)
	log.Printf("  Network: http://%s:%s\n", privateIP, port)

	log.Fatal(app.Run(address))
}
