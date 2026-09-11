package database

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"time"
)

//go:embed schema.sql
var schemaSQL string

// AutoMigrateDatabase सर्वर स्टार्ट होते ही schema.sql को रन करके टेबल्स बनाता है
func AutoMigrateDatabase(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("🔄 डेटाबेस स्कीमा माइग्रेशन शुरू हो रहा है...")

	_, err := db.ExecContext(ctx, schemaSQL)
	if err != nil {
		log.Printf("⚠️ माइग्रेशन त्रुटि: %v", err)
		return err
	}

	log.Println("✅ सभी डेटाबेस टेबल्स सफलतापूर्वक सत्यापित/निर्मित हो गईं!")
	return nil
}
