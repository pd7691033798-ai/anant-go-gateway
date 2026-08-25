package database

import (
	"database/sql"
	"log"
)

// InitCustomTables गो बैकएंड से अतिरिक्त सब्सक्रिप्शन और कोटे की टेबल्स स्वतः जोड़ता है
func InitCustomTables(db *sql.DB) error {
	query := `
	-- सब्सक्रिप्शन और प्लान टेबल (Demo, Basic, Pro, Family)
	CREATE TABLE IF NOT EXISTS subscriptions (
		id SERIAL PRIMARY KEY,
		whatsapp_number VARCHAR(20) REFERENCES users(phone),
		plan_tier VARCHAR(20) NOT NULL, -- DEMO, BASIC, PRO, FAMILY
		status VARCHAR(20) DEFAULT 'ACTIVE',
		batch_id VARCHAR(50),
		start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_date TIMESTAMP
	);

	-- दैनिक उपयोग और कोटे की निगरानी टेबल
	CREATE TABLE IF NOT EXISTS daily_usage (
		id SERIAL PRIMARY KEY,
		whatsapp_number VARCHAR(20) REFERENCES users(phone),
		usage_date DATE DEFAULT CURRENT_DATE,
		scans_used INT DEFAULT 0,
		qa_questions_used INT DEFAULT 0,
		UNIQUE (whatsapp_number, usage_date)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("❌ कस्टम टेबल्स बनाने में विफल: %v\n", err)
		return err
	}

	log.Println("✅ सब्सक्रिप्शन और डेली यूसेज की नई टेबल्स सफलतापूर्वक तैयार हो गई हैं!")
	return nil
}
