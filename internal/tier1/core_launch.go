package tier1

import (
	"database/sql"
	"fmt"
)

type CoreLaunchSuite struct {
	db *sql.DB
}

func NewCoreLaunchSuite(db *sql.DB) *CoreLaunchSuite {
	return &CoreLaunchSuite{db: db}
}

func (c *CoreLaunchSuite) GeneratePaymentQR(parentPhone, planType string, monthNumber int) (string, float64) {
	baseAmount := 399.0
	if planType == "PRO" {
		baseAmount = 699.0
	}
	finalAmount := baseAmount
	if monthNumber == 4 || monthNumber == 8 || monthNumber == 12 {
		finalAmount = baseAmount * 0.75
	}
	upiURI := fmt.Sprintf("upi://pay?pa=anantabhyas@upi&pn=AnantAbhyas&am=%.2f&cu=INR&tn=Abhyas_M%d_%s",
		finalAmount, monthNumber, parentPhone)
	return upiURI, finalAmount
}

func (c *CoreLaunchSuite) CheckDemoEligibility(phone string) (bool, int, error) {
	var daysUsed int
	var isActive bool
	query := `SELECT COALESCE(demo_days_used, 0), COALESCE(is_active, FALSE) FROM users WHERE phone = $1`
	err := c.db.QueryRow(query, phone).Scan(&daysUsed, &isActive)
	if err != nil {
		return true, 7, nil
	}
	if isActive {
		return true, 999, nil
	}
	if daysUsed >= 7 {
		return false, 0, nil
	}
	return true, 7 - daysUsed, nil
}
