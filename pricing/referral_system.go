package pricing

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type UserRole string

const (
	RoleParent  UserRole = "PARENT"
	RolePartner UserRole = "EXTERNAL_PARTNER" // दुकानदार, दोस्त, आदि
)

type ReferralAccount struct {
	RefCode         string    `json:"ref_code"`
	OwnerPhone      string    `json:"owner_phone"`
	OwnerName       string    `json:"owner_name"`
	Role            UserRole  `json:"role"`
	TotalReferred   int       `json:"total_referred"`
	PaidConversions int       `json:"paid_conversions"`
	EarnedBalance   float64   `json:"earned_balance"`   // अभिभावक के लिए बिल डिस्काउंट या पार्टनर के लिए कैश
	UPIHandle       string    `json:"upi_handle"`       // बाहरी व्यक्ति के पैसे भेजने के लिए
	CreatedAt       time.Time `json:"created_at"`
}

type ReferralSystemManager struct {
	accounts map[string]*ReferralAccount
	mu       sync.RWMutex
}

func NewReferralSystemManager() *ReferralSystemManager {
	return &ReferralSystemManager{
		accounts: make(map[string]*ReferralAccount),
	}
}

// 1. नया रेफरल लिंक जनरेट करना (अभिभावक या बाहरी व्यक्ति के लिए)
func (r *ReferralSystemManager) GenerateLink(phone, name string, role UserRole, upi string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	last4 := phone
	if len(phone) >= 4 {
		last4 = phone[len(phone)-4:]
	}

	var code string
	if role == RolePartner {
		cleanName := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
		if len(cleanName) > 3 {
			cleanName = cleanName[:3]
		}
		code = fmt.Sprintf("PARTNER_%s_%s", cleanName, last4)
	} else {
		// अभिभावक का कोड (पूरी तरह सुरक्षित)
		code = fmt.Sprintf("REF_%s", last4)
	}

	r.accounts[code] = &ReferralAccount{
		RefCode:       code,
		OwnerPhone:    phone,
		OwnerName:     name,
		Role:          role,
		UPIHandle:     upi,
		CreatedAt:     time.Now(),
	}

	return fmt.Sprintf("https://wa.me/9664006651?text=%s", code)
}

// 2. पेमेंट सक्सेसफुल होने पर रिवॉर्ड जोड़ना (नियम लागू करना)
func (r *ReferralSystemManager) ProcessSuccessfulPayment(refCode string, planPrice float64) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	cleanCode := strings.TrimSpace(strings.ToUpper(refCode))
	acc, exists := r.accounts[cleanCode]
	if !exists {
		return "अवैध रेफरल कोड।"
	}

	acc.PaidConversions++
	acc.TotalReferred++

	// यदि यह अभिभावक है: अधिकतम 2 रेफरल की सख्त सीमा (Cap of 2)
	if acc.Role == RoleParent {
		if acc.PaidConversions > 2 {
			return "⚠️ अभिभावक की अधिकतम 2 रेफरल की सीमा पूरी हो चुकी है। अतिरिक्त रेफरल पर वित्तीय छूट नहीं मिलेगी।"
		}
		// हर सफल रेफरल पर ₹100 की सीधी छूट
		acc.EarnedBalance += 100.0
		return fmt.Sprintf("✅ अभिभावक (%s) के खाते में ₹100 का बिल क्रेडिट जुड़ गया है।", acc.OwnerName)
	}

	// यदि यह बाहरी पार्टनर / दुकानदार है: चाहे बेसिक ले या प्रो, फिक्स ₹100 या तयशुदा कमिशन
	if acc.Role == RolePartner {
		commission := 100.0 // बेसिक या प्रो किसी पर भी ₹100 फ्लैट या प्लान अनुसार
		acc.EarnedBalance += commission
		return fmt.Sprintf("✅ पार्टनर (%s) के खाते में ₹%.0f का कैश कमीशन जमा हुआ। (UPI: %s)", acc.OwnerName, commission, acc.UPIHandle)
	}

	return "सफल।"
}

// 3. एडमिन डैशबोर्ड के लिए सभी खातों की सूची देखना
func (r *ReferralSystemManager) GetAdminReport() []ReferralAccount {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []ReferralAccount
	for _, acc := range r.accounts {
		list = append(list, *acc)
	}
	return list
}
