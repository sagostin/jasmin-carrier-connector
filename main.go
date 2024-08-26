package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"net"
	"os"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file. Using existing environment variables.")
	}

	// Initialize Loki client
	lokiClient := NewLokiClient(
		os.Getenv("LOKI_URL"),
		os.Getenv("LOKI_USERNAME"),
		os.Getenv("LOKI_PASSWORD"),
	)

	// Create custom logger
	logger := NewCustomLogger(lokiClient)

	app := fiber.New()

	gateway, err := NewSMSGateway(os.Getenv("MONGODB_URI"), logger)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to create SMS gateway: %v", err))
		os.Exit(1)
	}

	carriers, err := loadCarriers("carriers.json", logger)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to load carriers: %v", err))
		os.Exit(1)
	}
	gateway.Carriers = carriers

	for name, handler := range gateway.Carriers {
		inboundPath := fmt.Sprintf("/inbound/%s", name)
		//outboundPath := fmt.Sprintf("/outbound/%s", name)

		app.Post(inboundPath, func(c *fiber.Ctx) error {
			return handler.HandleInbound(c, gateway)
		})
		/*app.Post(outboundPath, func(c *fiber.Ctx) error {
			return handler.HandleOutbound(c, gateway)
		})*/
	}

	// Start server
	port := os.Getenv("WEB_LISTEN")
	if port == "" {
		port = "0.0.0.0:3000"
	}

	// load smpp clients / pbx credentials

	smppServer, _ = NewServer()
	smppServer.l, _ = net.Listen("tcp", os.Getenv("SMPP_LISTEN"))

	// server.ProcessSMS()

	// Add some example routes
	smppServer.AddRoute("1", "carrier", "twilio", carriers["twilio"])

	smppServer.Start()

	fmt.Printf("Starting SMPP server on %s\n", smppServer.Addr())

	err = app.Listen(port)
	if err != nil {
		log.Println(err)
	}
}
