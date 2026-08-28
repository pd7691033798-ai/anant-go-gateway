package pricing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type LoyaltyService struct {
	db *sql.DB
}

func NewLoyaltyService(db *sql.DB) *LoyaltyService {
	return &LoyaltyService{db: db}
}

// sanitizeLoyaltyPhone 10 अंकों का शुद्ध भारतीय मोबाइल नंबर निकालता है
func sanitizeLoyaltyPhone(phone string) string {
	clean := strings.TrimPrefix(phone, "+91")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	if len(clean) >= 10 {
		return clean[len(clean)-10:]
	}
	return clean
}

// CalculateLoyaltyFee आजीवन जब तक ग्राहक जुड़ा रहे, हर 4थे महीने (4, 8, 12, 16, 20... अनंत काल तक) 25% छात्रवृत्ति लागू करता है
func (l *LoyaltyService) CalculateLoyaltyFee(ctx context.Context, phone string, baseFee float64, tier PlanTier) (float64, string, int) {
	if tier == TierDemo {
		return baseFee, "7-दिवसीय डेमो परीक्षण सक्रिय", 0
	}

	cleanPhone := sanitizeLoyaltyPhone(phone)
	consecutiveMonths := 0

	if l.db != nil && len(cleanPhone) == 10 {
		dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		// डेटाबेस से लगातार भुगतान किए गए महीनों की संख्या प्राप्त करना
		query := `SELECT COALESCE(consecutive_paid_months, 0) FROM users WHERE phone = $1`
		err := l.db.QueryRowContext(dbCtx, query, cleanPhone).Scan(&consecutiveMonths)
		if err != nil {
			consecutiveMonths = 0
		}
	}

	// अगला महीना जिसके लिए फीस की गणना हो रही है
	nextMonth := consecutiveMonths + 1

	// गणितीय लॉजिक: हर 4थे महीने (nextMonth % 4 == 0) बिना किसी समय-सीमा के 25% छूट
	if nextMonth > 0 && nextMonth%4 == 0 {
		discountedFee := baseFee * 0.75
		scholarshipCycle := nextMonth / 4
		msg := fmt.Sprintf("🎉 25%% निरंतरता छात्रवृत्ति लागू (साइकिल #%d: महीना %d)", scholarshipCycle, nextMonth)
		return discountedFee, msg, nextMonth
	}

	monthsRemaining := 4 - (nextMonth % 4)
	msg := fmt.Sprintf("मानक मासिक शुल्क (महीना %d) - अगली 25%% छात्रवृत्ति के लिए %d माह शेष", nextMonth, monthsRemaining)
	return baseFee, msg, nextMonth
}

// ResetStreakOnDrop यदि ग्राहक सेवा छोड़ता है या नवीनीकरण नहीं करता, तो स्ट्रीक तुरंत 0 पर रीसेट हो जाती है
func (l *LoyaltyService) ResetStreakOnDrop(ctx context.Context, phone string) error {
	cleanPhone := sanitizeLoyaltyPhone(phone)
	if l.db == nil || len(cleanPhone) != 10 {
		return fmt.Errorf("अमान्य डेटाबेस कनेक्शन या फ़ोन नंबर")
	}

	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		UPDATE users 
		SET consecutive_paid_months = 0,
		    updated_at = NOW() 
		WHERE phone = $1`

	_, err := l.db.ExecContext(dbCtx, query, cleanPhone)
	return err
}

// RecordSuccessfulPayment सफल भुगतान होने पर लगातार महीनों का काउंटर +1 करता है
func (l *LoyaltyService) RecordSuccessfulPayment(ctx context.Context, phone string) error {
	cleanPhone := sanitizeLoyaltyPhone(phone)
	if l.db == nil || len(cleanPhone) != 10 {
		return fmt.Errorf("अमान्य डेटाबेस कनेक्शन या फ़ोन नंबर")
	}

	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		UPDATE users 
		SET consecutive_paid_months = COALESCE(consecutive_paid_months, 0) + 1,
		    updated_at = NOW() 
		WHERE phone = $1`

	_, err := l.db.ExecContext(dbCtx, query, cleanPhone)
	return err
}
