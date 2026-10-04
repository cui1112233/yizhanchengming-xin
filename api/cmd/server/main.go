package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/app"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/provider121"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("QIANTIE_MYSQL_DSN")
	if dsn == "" {
		log.Fatal("QIANTIE_MYSQL_DSN is required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}

	fetcher, err := provider121.NewClient(provider121.Config{})
	if err != nil {
		log.Fatalf("create 121 client: %v", err)
	}

	addr := os.Getenv("QIANTIE_GO_LISTEN_ADDR")
	if addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, app.NewHandler(db, fetcher, nil, nil)); err != nil {
		log.Fatal(err)
	}
}
