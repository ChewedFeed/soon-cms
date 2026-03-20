package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/chewedfeed/soon-cms/internal/auth"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: create-admin <username> <password>\n")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	connStr := os.Getenv("RDS_URL")
	if connStr == "" {
		logs.Fatalf("RDS_URL environment variable is required")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		logs.Fatalf("Failed to hash password: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logs.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO admin_users (username, password_hash) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET password_hash = $2",
		username, hash,
	)
	if err != nil {
		logs.Fatalf("Failed to create admin user: %v", err)
	}

	logs.Logf("Admin user '%s' created/updated successfully", username)
}
