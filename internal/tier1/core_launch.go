package tier1

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type CoreLaunchSuite struct {
	db *sql.DB
}

func NewCoreLaunchSuite(db *sql.DB) *CoreLaunchSuite {
	return &CoreLaunchSuite{db: db}
}

// sanitizePhone 10 अंकों का शुद्ध नंबर निकालता है
func sanitizePhone(phone string) string {
	clean := strings.TrimPrefix(phone, "+91")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	if len(clean) >= 10 {
		return clean[len(clean)-10:]
	}
	return clean
}

func (c *CoreLaunchSuite) GeneratePaymentQR(parentPhone, planType string, monthNumber int) (string, float64) {
	phone := sanitizePhone(parentPhone)
	plan := strings.ToUpper(strings.TrimSpace(planType))

	var baseAmount float64
	switch plan {
	case "PRO":
		baseAmount = 699.0
	case "FAMILY":
		baseAmount = 1099.0
	case "UNLIMITED_FAMILY":
		baseAmount = 1499.0
	default:
		baseAmount = 399.0 // BASIC Plan
	}

	finalAmount := baseAmount
	// 4थे, 8वें और 12वें महीने पर 25% छूट
	if monthNumber == 4 || monthNumber == 8 || monthNumber == 12 {
		finalAmount = baseAmount * 0.75
	}

	payeeVPA := "anantabhyas@upi"
	payeeName := "Anant Abhyas"
	transactionNote := fmt.Sprintf("Abhyas_M%d_%s", monthNumber, phone)

	// सुरक्षित UPI URL फॉर्मेट
	upiURI := fmt.Sprintf(
		"upi://pay?pa=%s&pn=%s&am=%.2f&cu=INR&tn=%s",
		url.QueryEscape(payeeVPA),
		url.QueryEscape(payeeName),
		finalAmount,
		url.QueryEscape(transactionNote),
	)

	return upiURI, finalAmount
}

func (c *CoreLaunchSuite) CheckDemoEligibility(phone string) (bool, int, error) {
	cleanPhone := sanitizePhone(phone)
	if len(cleanPhone) != 10 {
		return false, 0, fmt.Errorf("अमान्य फ़ोन नंबर: %s", phone)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var daysUsed int
	var isActive bool
	query := `SELECT COALESCE(demo_days_used, 0), COALESCE(is_active, FALSE) FROM users WHERE phone = $1`
	err := c.db.QueryRowContext(ctx, query, cleanPhone).Scan(&daysUsed, &isActive)

	if err != nil {
		if err == sql.ErrNoRows {
			// बिल्कुल नया यूज़र - 7 दिन का फ़्री डेमो उपलब्ध
			return true, 7, nil
		}
		// असली डेटाबेस एरर
		return false, 0, fmt.Errorf("डेमो पात्रता जांच विफल: %w", err)
	}

	// यदि एक्टिव पेड सब्सक्रिप्शन है
	if isActive {
		return true, 999, nil
	}

	// यदि 7 दिन पूरे हो चुके हैं
	if daysUsed >= 7 {
		return false, 0, nil
	}

	return true, 7 - daysUsed, nil
}
