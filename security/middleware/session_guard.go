package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"your_project_name/pricing"
	"your_project_name/security" // अपने प्रोजेक्ट के गो मॉड्यूल नाम के अनुसार रखें
)

type contextKey string

const (
	StudentUIDKey contextKey = "studentUID"
	ParentUIDKey  contextKey = "parentUID"
	DeviceIDKey   contextKey = "deviceID"
)

type SessionGuardMiddleware struct {
	deviceGuard    *security.DeviceGuardService
	familyGuardian *pricing.FamilyGuardianService
}

func NewSessionGuardMiddleware(deviceGuard *security.DeviceGuardService, familyGuardian *pricing.FamilyGuardianService) *SessionGuardMiddleware {
	return &SessionGuardMiddleware{
		deviceGuard:    deviceGuard,
		familyGuardian: familyGuardian,
	}
}

// respondJSON मानकीकृत JSON एरर संदेश भेजने का हेल्पर फ़ंक्शन
func respondJSON(w http.ResponseWriter, statusCode int, message string, action security.DeviceAction) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"action":  action,
		"message": message,
	})
}

// RequireValidDeviceApp ऐप से आने वाले API कॉल्स के लिए डिवाइस वैधता और सिंगल एक्टिव सेशन की जाँच करता है
func (m *SessionGuardMiddleware) RequireValidDeviceApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		studentUID := strings.TrimSpace(r.Header.Get("X-Student-UID"))
		parentUID := strings.TrimSpace(r.Header.Get("X-Parent-UID"))
		deviceID := strings.TrimSpace(r.Header.Get("X-Device-ID"))

		if deviceID == "" {
			respondJSON(w, http.StatusBadRequest, "अमान्य डिवाइस पहचान (Device ID missing)", security.ActionForceLogout)
			return
		}

		// 1. यदि यह फैमिली प्लान का मास्टर अकाउंट सत्र है, तो फैमिली डिवाइस गार्ड चलाएँ
		if parentUID != "" && m.familyGuardian != nil {
			isValid, err := m.familyGuardian.ValidateDeviceSession(r.Context(), parentUID, deviceID)
			if err != nil || !isValid {
				respondJSON(w, http.StatusUnauthorized, "सत्र समाप्त: आपका फैमिली खाता किसी अन्य डिवाइस पर सक्रिय किया गया है।", security.ActionForceLogout)
				return
			}
		}

		// 2. स्टुडेंट लेवल डिवाइस और सेशन स्टेट की जाँच
		action, msg, err := m.deviceGuard.HandleAppLaunch(r.Context(), studentUID, deviceID)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, "सर्वर सत्यापन में त्रुटि आई। कृपया पुनः प्रयास करें।", security.ActionForceLogout)
			return
		}

		// यदि ऐप किसी नए/अनधिकृत फोन में खोला गया है (Shared APK Protection)
		if action == security.ActionShowOnboarding {
			respondJSON(w, http.StatusUnauthorized, msg, security.ActionShowOnboarding)
			return
		}

		if action == security.ActionForceLogout {
			respondJSON(w, http.StatusUnauthorized, "यह खाता किसी अन्य डिवाइस पर खोला गया है।", security.ActionForceLogout)
			return
		}

		// 3. सत्यापित डेटा को रिक्वेस्ट कॉन्टेक्स्ट में इंजेक्ट करें
		ctx := context.WithValue(r.Context(), StudentUIDKey, studentUID)
		ctx = context.WithValue(ctx, DeviceIDKey, deviceID)
		if parentUID != "" {
			ctx = context.WithValue(ctx, ParentUIDKey, parentUID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateExpiringWebSession Chrome / ब्राउज़र लिंक की समय-सीमा और डिजिटल सिग्नेचर जाँचता है
func (m *SessionGuardMiddleware) ValidateExpiringWebSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		uid := strings.TrimSpace(query.Get("uid"))
		exp := strings.TrimSpace(query.Get("exp"))
		sig := strings.TrimSpace(query.Get("sig"))

		if uid == "" || exp == "" || sig == "" {
			respondJSON(w, http.StatusBadRequest, "अमान्य या अधूरा अभ्यास लिंक। कृपया WhatsApp से नया लिंक प्राप्त करें।", security.ActionForceLogout)
			return
		}

		// लिंक सिग्नेचर और 60-मिनट मियाद की जाँच
		isValid, validationMsg := m.deviceGuard.VerifyExpiringWebLink(uid, exp, sig)
		if !isValid {
			respondJSON(w, http.StatusForbidden, validationMsg, security.ActionForceLogout)
			return
		}

		ctx := context.WithValue(r.Context(), StudentUIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

