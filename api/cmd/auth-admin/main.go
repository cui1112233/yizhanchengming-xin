package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("QIANTIE_MYSQL_DSN")
	if dsn == "" {
		log.Fatal("QIANTIE_MYSQL_DSN is required")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil { log.Fatalf("open mysql: %v", err) }
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil { log.Fatalf("ping mysql: %v", err) }

	store := authn.NewMySQLStore(db)
	action := strings.TrimSpace(os.Getenv("QIANTIE_AUTH_ADMIN_ACTION"))
	username := strings.TrimSpace(os.Getenv("QIANTIE_AUTH_ADMIN_USERNAME"))
	password := os.Getenv("QIANTIE_AUTH_ADMIN_PASSWORD")

	switch action {
	case "bootstrap":
		created, err := store.EnsureInitialAdmin(ctx, username, password)
		if err != nil { log.Fatal(err) }
		fmt.Printf("bootstrap created=%t\n", created)
	case "set-password":
		if err := store.SetUserPassword(ctx, username, password); err != nil { log.Fatal(err) }
		fmt.Println("password updated")
	case "disable":
		if err := store.SetUserActive(ctx, username, false); err != nil { log.Fatal(err) }
		fmt.Println("user disabled")
	case "enable":
		if err := store.SetUserActive(ctx, username, true); err != nil { log.Fatal(err) }
		fmt.Println("user enabled")
	default:
		log.Fatal("QIANTIE_AUTH_ADMIN_ACTION must be bootstrap, set-password, disable, or enable")
	}
}
