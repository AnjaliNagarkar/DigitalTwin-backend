package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() *sql.DB {
	dsn := "playground:Pl@Ygr0und@tcp(10.15.20.235:3306)/ivdp_db?parseTime=true"

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	if err := conn.Ping(); err != nil {
		log.Fatalf("Failed to reach DB: %v", err)
	}

	fmt.Println("Connected to MySQL successfully")
	return conn
}
