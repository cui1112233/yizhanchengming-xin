package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/app"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
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

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := db.PingContext(pingCtx); err != nil {
		pingCancel()
		log.Fatalf("ping mysql: %v", err)
	}
	pingCancel()

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 15*time.Second)
	created, err := authn.NewMySQLStore(db).EnsureInitialAdmin(
		bootstrapCtx,
		strings.TrimSpace(os.Getenv("QIANTIE_BOOTSTRAP_ADMIN_USERNAME")),
		os.Getenv("QIANTIE_BOOTSTRAP_ADMIN_PASSWORD"),
	)
	bootstrapCancel()
	if err != nil {
		log.Fatalf("initialize auth administrator: %v", err)
	}
	if created {
		log.Printf("initial auth administrator created; remove QIANTIE_BOOTSTRAP_ADMIN_PASSWORD from the runtime environment after first startup")
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
