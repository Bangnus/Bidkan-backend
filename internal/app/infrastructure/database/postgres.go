package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// ทดสอบว่าต่อได้จริง
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Production Tuning
	db.SetMaxOpenConns(25)                 // จำนวน Connection สูงสุดที่เปิดพร้อมกันได้
	db.SetMaxIdleConns(25)                 // จำนวน Connection ที่เปิดรอไว้
	db.SetConnMaxLifetime(5 * time.Minute) // อายุสูงสุดของ Connection

	return db
}