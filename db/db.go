package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() *sql.DB {
	dsn := "playground:Pl@Ygr0und@tcp(10.15.20.235:3306)/ivdp_db?parseTime=true"

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(5 * time.Minute)
	conn.SetConnMaxIdleTime(2 * time.Minute)

	if err := conn.Ping(); err != nil {
		log.Fatalf("Failed to reach DB: %v", err)
	}

	fmt.Println("Connected to MySQL successfully")
	return conn
}
