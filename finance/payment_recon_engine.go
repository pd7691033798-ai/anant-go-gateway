package finance

import (
	"errors"
	"fmt"
	"time"
)

type PaymentMode string
type PaymentStatus string

const (
	ModeAutoDebitUPI PaymentMode = "AUTO_DEBIT_UPI"
	ModeManualPay    PaymentMode = "MANUAL_UPI_OR_CASH"

	StatusInitiated PaymentStatus = "INITIATED"
	StatusSettled   PaymentStatus = "SETTLED_TO_ADMIN_BANK"
	StatusInGrace   PaymentStatus = "GRACE_PERIOD_ACTIVE"
	StatusFailed    PaymentStatus = "FAILED"
)

type PaymentRecord struct {
	TransactionID   string
	ParentID        string
	AmountINR       int
	Mode            PaymentMode
	Status          PaymentStatus
	BankReferenceNo string    // बैंक UTR या सेटलमेंट ID
	SettledAt       time.Time
	GraceUntil      time.Time
}

type BillingReconEngine struct {
	ledger map[string]*PaymentRecord
}

func NewBillingReconEngine() *BillingReconEngine {
	return &BillingReconEngine{
		ledger: make(map[string]*PaymentRecord),
	}
}

// ProcessSubscriptionAccess: तय करना कि बच्चे को अभ्यास करने देना है या नहीं
func (e *BillingReconEngine) ProcessSubscriptionAccess(parentID string, mode PaymentMode, amount int, hasImmediateFunds bool) (*PaymentRecord, error) {
	txID := fmt.Sprintf("TX_%d_%s", time.Now().Unix(), parentID)

	// अगर माता-पिता के पास तुरंत पैसे नहीं हैं और वे मैनुअल मोड चुनते हैं
	if !hasImmediateFunds && mode == ModeManualPay {
		record := &PaymentRecord{
			TransactionID: txID,
			ParentID:      parentID,
			AmountINR:     amount,
			Mode:          ModeManualPay,
			Status:        StatusInGrace,
			GraceUntil:    time.Now().Add(1 * 24 * time.Hour), // 1 दिन की पढ़ाई चालू (ग्रेस पीरियड)
		}
		e.ledger[parentID] = record
		return record, nil
	}

	// सामान्य ऑटो-पे या तुरंत भुगतान
	record := &PaymentRecord{
		TransactionID: txID,
		ParentID:      parentID,
		AmountINR:     amount,
		Mode:          mode,
		Status:        StatusInitiated,
	}
	e.ledger[parentID] = record
	return record, nil
}

// ConfirmAdminBankSettlement: जब पेमेंट गेटवे/बैंक से पैसा एडमिन खाते में जमा होने का वेबहुक आए
func (e *BillingReconEngine) ConfirmAdminBankSettlement(parentID, bankUTR string) error {
	record, exists := e.ledger[parentID]
	if !exists {
		return errors.New("लेनदेन रिकॉर्ड नहीं मिला")
	}

	record.Status = StatusSettled
	record.BankReferenceNo = bankUTR
	record.SettledAt = time.Now()
	return nil
}

// IsAccessAllowed: क्या बच्चे का 15-मिनट CBT खुला रहेगा?
func (e *BillingReconEngine) IsAccessAllowed(parentID string) bool {
	record, exists := e.ledger[parentID]
	if !exists {
		return false
	}
	if record.Status == StatusSettled {
		return true
	}
	if record.Status == StatusInGrace && time.Now().Before(record.GraceUntil) {
		return true // ग्रेस पीरियड में पढ़ाई नहीं रुकेगी
	}
	return false
}
