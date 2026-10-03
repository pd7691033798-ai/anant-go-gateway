package finance

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// MandateStatus ऑटो-पे की वर्तमान स्थिति को दर्शाता है
type MandateStatus string

const (
	MandatePending MandateStatus = "PENDING_MANDATE_APPROVAL"
	MandateActive  MandateStatus = "MANDATE_ACTIVE_AUTOPAY"
	MandateFailed  MandateStatus = "AUTOPAY_FAILED"
	MandateRevoked MandateStatus = "MANDATE_CANCELLED" // जब पेरेंट या एडमिन खुद बंद करे
)

type AutoPaySubscription struct {
	SubscriptionID string
	ParentID       string
	MonthlyAmount  int
	MandateStatus  MandateStatus
	NextBillingAt  time.Time
	GraceExpiry    time.Time
	LastSettledUTR string
	IsAccessOpen   bool
}

type AutoPayManager struct {
	db *sql.DB
	// म्यूटेक्स (sync.RWMutex) हटा दिया गया है, क्योंकि अब सारा लोड डेटाबेस के ट्रांजैक्शन हैंडल करेंगे।
}

// NewAutoPayManager डेटाबेस कनेक्शन के साथ इंजन को इनिशियलाइज करता है
func NewAutoPayManager(db *sql.DB) *AutoPayManager {
	return &AutoPayManager{db: db}
}

// SetupMandate: ऑनबोर्डिंग पर UPI ऑटो-पे ऑथराइज कराना
func (m *AutoPayManager) SetupMandate(parentID string, amount int) (*AutoPaySubscription, error) {
	if m.db == nil {
		return nil, errors.New("database connection is required")
	}

	// ग्लोबल UTC टाइम ताकि दुनिया के किसी भी सर्वर पर समय एक ही रहे
	now := time.Now().UTC()
	subID := fmt.Sprintf("SUB_AUTOPAY_%d_%s", now.Unix(), parentID)
	
	nextBilling := now.Add(30 * 24 * time.Hour)
	graceExpiry := nextBilling.Add(24 * time.Hour) // केवल 1 दिन का ग्रेस पीरियड

	sub := &AutoPaySubscription{
		SubscriptionID: subID,
		ParentID:       parentID,
		MonthlyAmount:  amount,
		MandateStatus:  MandatePending,
		NextBillingAt:  nextBilling,
		GraceExpiry:    graceExpiry,
		IsAccessOpen:   false, // जब तक बैंक से सक्सेस न आए, तब तक एक्सेस बंद रहेगा
	}

	// ON CONFLICT: अगर पेरेंट दो बार बटन दबा दे, तो डेटाबेस क्रैश नहीं होगा
	query := `INSERT INTO autopay_subscriptions 
		(subscription_id, parent_id, monthly_amount, mandate_status, next_billing_at, grace_expiry, is_access_open, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (parent_id) DO UPDATE 
		SET mandate_status = EXCLUDED.mandate_status, next_billing_at = EXCLUDED.next_billing_at`
	
	_, err := m.db.Exec(query, sub.SubscriptionID, sub.ParentID, sub.MonthlyAmount, sub.MandateStatus, sub.NextBillingAt, sub.GraceExpiry, sub.IsAccessOpen, now)
	if err != nil {
		return nil, fmt.Errorf("failed to save mandate: %v", err)
	}
	
	return sub, nil
}

// HandleAutoDebitWebhook: 10M स्केल के लिए डेटाबेस ट्रांजैक्शन (Transaction) और रो-लॉकिंग (Row-Locking) के साथ
func (m *AutoPayManager) HandleAutoDebitWebhook(parentID, bankUTR string, isSuccess bool) error {
	if m.db == nil {
		return errors.New("database connection is nil")
	}

	now := time.Now().UTC()

	// 1. ट्रांजैक्शन शुरू करें (रेस कंडीशन रोकने के लिए)
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	// सुरक्षा: अगर फंक्शन बीच में फेल हो जाए, तो जो भी बदलाव हुए हैं उन्हें वापस ले लो (Rollback)
	defer tx.Rollback() 

	var currentStatus string
	var currentGraceExpiry time.Time

	// 2. FOR UPDATE: इस यूज़र की रो (row) को लॉक कर दो ताकि जब तक हम अपडेट न कर लें, कोई और वेबहुक इसे न पढ़ सके
	err = tx.QueryRow("SELECT mandate_status, grace_expiry FROM autopay_subscriptions WHERE parent_id = $1 FOR UPDATE", parentID).Scan(&currentStatus, &currentGraceExpiry)
	if err != nil {
		return errors.New("subscription not found for this parent")
	}

	var newStatus MandateStatus
	var nextBilling time.Time
	var graceExpiry time.Time
	var isAccessOpen bool

	// 3. पेमेंट स्टेटस के आधार पर लॉजिक
	if isSuccess {
		newStatus = MandateActive
		nextBilling = now.Add(30 * 24 * time.Hour)
		graceExpiry = nextBilling.Add(24 * time.Hour)
		isAccessOpen = true
	} else {
		newStatus = MandateFailed
		nextBilling = now
		
		// लूपहोल फिक्स: अगर यह पहले से ही Failed है, तो नया 24 घंटा नहीं देंगे, पुराना वाला ही रखेंगे
		if currentStatus == string(MandateFailed) {
			graceExpiry = currentGraceExpiry 
		} else {
			graceExpiry = now.Add(24 * time.Hour) 
		}
		isAccessOpen = true 
	}

	// 4. अपडेट क्वेरी चलाएं
	updateQuery := `UPDATE autopay_subscriptions 
		SET mandate_status = $1, next_billing_at = $2, grace_expiry = $3, last_settled_utr = $4, is_access_open = $5 
		WHERE parent_id = $6`
	_, err = tx.Exec(updateQuery, newStatus, nextBilling, graceExpiry, bankUTR, isAccessOpen, parentID)
	if err != nil {
		return fmt.Errorf("failed to update autopay status: %v", err)
	}

	// 5. सब कुछ सही रहा, तो डेटाबेस में हमेशा के लिए सेव (Commit) कर दो
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	if !isSuccess {
		return errors.New("AUTOPAY_FAILED: पेमेंट असफल! केवल 1 दिन का ग्रेस पीरियड सक्रिय है।")
	}
	return nil
}

// CancelMandate: पेरेंट या कस्टमर सपोर्ट के कहने पर सब्सक्रिप्शन तुरंत ब्लॉक करने का फंक्शन
func (m *AutoPayManager) CancelMandate(parentID string) error {
	if m.db == nil {
		return errors.New("database connection is nil")
	}

	query := `UPDATE autopay_subscriptions SET mandate_status = $1, is_access_open = false WHERE parent_id = $2`
	_, err := m.db.Exec(query, MandateRevoked, parentID)
	if err != nil {
		return fmt.Errorf("failed to cancel mandate: %v", err)
	}
	
	return nil
}

// CanChildPractice: क्या बच्चे का अभ्यास खुला है?
func (m *AutoPayManager) CanChildPractice(parentID string) bool {
	if m.db == nil {
		return false
	}

	var isAccessOpen bool
	var nextBilling, graceExpiry time.Time
	var mandateStatus string

	err := m.db.QueryRow(`SELECT is_access_open, next_billing_at, grace_expiry, mandate_status FROM autopay_subscriptions WHERE parent_id = $1`, parentID).Scan(&isAccessOpen, &nextBilling, &graceExpiry, &mandateStatus)
	if err != nil {
		return false
	}

	// सुरक्षा: अगर कैंसल कर दिया है, तो तुरंत ब्लॉक
	if mandateStatus == string(MandateRevoked) {
		return false
	}

	now := time.Now().UTC()
	
	// अगर एक्सेस ओपन है और समय सीमा या 1 दिन के ग्रेस पीरियड के अंदर है, तभी चलने दो
	if isAccessOpen && (now.Before(nextBilling) || now.Before(graceExpiry)) {
		return true
	}

	// समय समाप्त, खेल खत्म!
	return false
}
