package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("डेटाबेस कनेक्शन एरर: %v", err)
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("⚠️ डेटाबेस पिंग विफल (ऑफलाइन टेस्ट मोड सक्रिय): %v\n", err)
	} else {
		fmt.Println("✅ PostgreSQL डेटाबेस सफलतापूर्वक कनेक्ट हुआ।")
	}

	return db
}

// ConnectPostgres: main.go के साथ कम्पैटिबिलिटी के लिए
func ConnectPostgres(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return db, fmt.Errorf("डेटाबेस पिंग विफल: %w", err)
	}
	return db, nil
}
