type AccessStatus string

const (
	StatusActivePaid   AccessStatus = "ACTIVE_PAID"   // पेड सब्सक्रिप्शन एक्टिव
	StatusActiveDemo   AccessStatus = "ACTIVE_DEMO"   // दिन 1-7 (फ्री डेमो)
	StatusGracePeriod  AccessStatus = "GRACE_PERIOD"  // दिन 8-10 (3 दिन का ग्रेस पीरियड)
	StatusExpired      AccessStatus = "EXPIRED"       // 10 दिन पूरे - सेवा बंद
)

// CheckAccessStatus 7 दिन डेमो + 3 दिन ग्रेस पीरियड की जांच करता है
func (c *CoreLaunchSuite) CheckAccessStatus(phone string) (AccessStatus, int, error) {
	cleanPhone := sanitizePhone(phone)
	if len(cleanPhone) != 10 {
		return StatusExpired, 0, fmt.Errorf("अमान्य फ़ोन नंबर: %s", phone)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var daysUsed int
	var isActive bool
	query := `SELECT COALESCE(demo_days_used, 0), COALESCE(is_active, FALSE) FROM users WHERE phone = $1`
	err := c.db.QueryRowContext(ctx, query, cleanPhone).Scan(&daysUsed, &isActive)

	if err != nil {
		if err == sql.ErrNoRows {
			// बिल्कुल नया छात्र: 7 दिन डेमो
			return StatusActiveDemo, 7, nil
		}
		return StatusExpired, 0, fmt.Errorf("स्टेटस जांच विफल: %w", err)
	}

	// 1. यदि पेड प्लान चालू है
	if isActive {
		return StatusActivePaid, 999, nil
	}

	// 2. दिन 1 से 7: फ़्री डेमो
	if daysUsed < 7 {
		remainingDemoDays := 7 - daysUsed
		return StatusActiveDemo, remainingDemoDays, nil
	}

	// 3. दिन 8 से 10 (कुल 10 दिन तक): 3 दिन का ग्रेस पीरियड
	if daysUsed >= 7 && daysUsed < 10 {
		remainingGraceDays := 10 - daysUsed
		return StatusGracePeriod, remainingGraceDays, nil
	}

	// 4. 10 दिन पूरे: ट्रायल व ग्रेस समाप्त
	return StatusExpired, 0, nil
}
