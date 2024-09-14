package main

import (
	_ "embed"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"matbm.net/geonow/handlers"
	"matbm.net/geonow/ratelimit"
	"net/http"
)

func main() {
	// Initialize libvips
	vips.Startup(nil)
	defer vips.Shutdown()

	// Start rate limit routine
	go ratelimit.CleanRateLimits()

	// Configure the handlers
	http.HandleFunc("/", handlers.ImageHandler)
	http.HandleFunc("/r", handlers.RedirectorHandler)

	// Start the webserver
	serveAddr := ":8080"
	log.Printf("Server is running at %s", serveAddr)
	err := http.ListenAndServe(serveAddr, nil)
	if err != nil {
		log.Printf("Failed to serve at %s", serveAddr)
	}
}
