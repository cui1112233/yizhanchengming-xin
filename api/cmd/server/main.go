package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, httpapi.NewHandler()); err != nil {
		log.Fatal(err)
	}
}
