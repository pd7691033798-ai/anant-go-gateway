package pricing

import (
	"database/sql"
	"fmt"
)

type LoyaltyService struct {
	db *sql.DB
}

func NewLoyaltyService(db *sql.DB) *LoyaltyService {
	return &LoyaltyService{db: db}
}

func (l *LoyaltyService) CalculateLoyaltyFee(phone string, baseFee float64, tier PlanTier) (float64, string) {
	if tier == TierDemo {
		return baseFee, "7-दिवसीय डेमो परीक्षण सक्रिय"
	}

	var consecutiveMonths int
	err := l.db.QueryRow(`SELECT consecutive_paid_months FROM users WHERE phone = $1`, phone).Scan(&consecutiveMonths)
	if err != nil {
		consecutiveMonths = 1
	}

	nextMonth := consecutiveMonths + 1

	if nextMonth == 4 || nextMonth == 8 || nextMonth == 12 {
		discountedFee := baseFee * 0.75
		scholarshipNum := nextMonth / 4
		return discountedFee, fmt.Sprintf("25%% निरंतरता छात्रवृत्ति लागू (%d/3)", scholarshipNum)
	}

	return baseFee, "मानक मासिक शुल्क"
}
