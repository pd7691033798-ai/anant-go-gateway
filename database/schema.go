package database

import (
	"database/sql"
	"log"
)

// InitSchema डेटाबेस में आवश्यक टेबल्स (जैसे यूज़र, सब्सक्रिप्शन और कोटा) स्वतः बनाता है
func InitSchema(db *sql.DB) error {
	query := `
	-- 1. यूज़र टेबल (WhatsApp नंबर आधारित)
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		whatsapp_number VARCHAR(20) UNIQUE NOT NULL,
		name VARCHAR(100),
		language VARCHAR(20) DEFAULT 'hi',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 2. सब्सक्रिप्शन और प्लान टेबल (Demo, Basic, Pro, Family)
	CREATE TABLE IF NOT EXISTS subscriptions (
		id SERIAL PRIMARY KEY,
		whatsapp_number VARCHAR(20) REFERENCES users(whatsapp_number),
		plan_tier VARCHAR(20) NOT NULL, -- DEMO, BASIC, PRO, FAMILY
		status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, EXPIRED
		batch_id VARCHAR(50),
		start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_date TIMESTAMP
	);

	-- 3. दैनिक उपयोग और कोटे की निगरानी टेबल (स्कैन और AI सवाल)
	CREATE TABLE IF NOT EXISTS daily_usage (
		id SERIAL PRIMARY KEY,
		whatsapp_number VARCHAR(20) REFERENCES users(whatsapp_number),
		usage_date DATE DEFAULT CURRENT_DATE,
		scans_used INT DEFAULT 0,
		qa_questions_used INT DEFAULT 0,
		UNIQUE (whatsapp_number, usage_date)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("❌ डेटाबेस टेबल बनाने में विफल: %v\n", err)
		return err
	}

	log.Println("✅ सभी डेटाबेस टेबल्स (Users, Subscriptions, Usage) सफलतापूर्वक तैयार हो गई हैं!")
	return nil
}
