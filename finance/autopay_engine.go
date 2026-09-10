package finance

import (
	"errors"
	"fmt"
	"time"
)

type MandateStatus string

const (
	MandatePending MandateStatus = "PENDING_MANDATE_APPROVAL"
	MandateActive  MandateStatus = "MANDATE_ACTIVE_AUTOPAY"
	MandateFailed  MandateStatus = "AUTOPAY_FAILED"
	MandateRevoked MandateStatus = "MANDATE_CANCELLED"
)

type AutoPaySubscription struct {
	SubscriptionID string
	ParentID       string
	MonthlyAmount  int           // जैसे ₹1,198 या ₹1,898
	MandateStatus  MandateStatus
	NextBillingAt  time.Time
	LastSettledUTR string
	IsAccessOpen   bool
}

type AutoPayManager struct {
	subscriptions map[string]*AutoPaySubscription
}

func NewAutoPayManager() *AutoPayManager {
	return &AutoPayManager{
		subscriptions: make(map[string]*AutoPaySubscription),
	}
}

// SetupMandate: ऑनबोर्डिंग पर ही माता-पिता से UPI ऑटो-पे ऑथराइज कराना
func (m *AutoPayManager) SetupMandate(parentID string, amount int) *AutoPaySubscription {
	subID := fmt.Sprintf("SUB_AUTOPAY_%d_%s", time.Now().Unix(), parentID)
	sub := &AutoPaySubscription{
		SubscriptionID: subID,
		ParentID:       parentID,
		MonthlyAmount:  amount,
		MandateStatus:  MandatePending,
		NextBillingAt:  time.Now().Add(30 * 24 * time.Hour),
		IsAccessOpen:   false,
	}
	m.subscriptions[parentID] = sub
	return sub
}

// HandleAutoDebitWebhook: जब बैंक से ऑटो-पे कटने का कन्फर्मेशन आए
func (m *AutoPayManager) HandleAutoDebitWebhook(parentID, bankUTR string, isSuccess bool) error {
	sub, exists := m.subscriptions[parentID]
	if !exists {
		return errors.New("सब्सक्रिप्शन रिकॉर्ड नहीं मिला")
	}

	if isSuccess {
		sub.MandateStatus = MandateActive
		sub.LastSettledUTR = bankUTR
		sub.NextBillingAt = time.Now().Add(30 * 24 * time.Hour)
		sub.IsAccessOpen = true // ऑटो-पे सफल होते ही 30 दिन का अभ्यास अनलॉक
		return nil
	}

	// अगर ऑटो-पे बैंक की तरफ से फेल हो जाए (खाते में बैलेंस न होना)
	sub.MandateStatus = MandateFailed
	sub.IsAccessOpen = false // नो ग्रेस पीरियड - तुरंत ब्लॉक
	return errors.New("AUTOPAY_FAILED: खाते से ऑटो-पे नहीं कट सका, अभ्यास अस्थायी रूप से रुका")
}

// CanChildPractice: क्या बच्चे का 15-मिनट CBT खुला है?
func (m *AutoPayManager) CanChildPractice(parentID string) bool {
	sub, exists := m.subscriptions[parentID]
	if !exists {
		return false
	}
	// केवल वही बच्चा पढ़ सकता है जिसका ऑटो-पे एक्टिव है और वैलिडिटी बची है
	return sub.IsAccessOpen && time.Now().Before(sub.NextBillingAt)
}
