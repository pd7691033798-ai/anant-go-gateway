package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"your_project_name/security" // प्रोजेक्ट के अनुसार इम्पोर्ट पाथ सेट करें
)

type contextKey string

const (
	StudentUIDKey contextKey = "studentUID"
	DeviceIDKey   contextKey = "deviceID"
)

type SessionGuardMiddleware struct {
	deviceGuard *security.DeviceGuardService
}

func NewSessionGuardMiddleware(deviceGuard *security.DeviceGuardService) *SessionGuardMiddleware {
	return &SessionGuardMiddleware{
		deviceGuard: deviceGuard,
	}
}

// RespondJSON मानकीकृत JSON एरर संदेश भेजने का हेल्पर फ़ंक्शन
func respondJSON(w http.ResponseWriter, statusCode int, message string, action security.DeviceAction) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"action":  action,
		"message": message,
	})
}

// RequireValidDeviceApp ऐप से आने वाले API कॉल्स के लिए डिवाइस वैधता की जाँच करता है
func (m *SessionGuardMiddleware) RequireValidDeviceApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		studentUID := strings.TrimSpace(r.Header.Get("X-Student-UID"))
		deviceID := strings.TrimSpace(r.Header.Get("X-Device-ID"))

		// डिवाइस और सेशन स्टेट की जाँच
		action, msg, err := m.deviceGuard.HandleAppLaunch(r.Context(), studentUID, deviceID)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, "सर्वर सत्यापन में त्रुटि आई। कृपया पुनः प्रयास करें।", security.ActionForceLogout)
			return
		}

		// यदि ऐप किसी नए फोन में खोला गया है (Shared APK)
		if action == security.ActionShowOnboarding {
			respondJSON(w, http.StatusUnauthorized, msg, security.ActionShowOnboarding)
			return
		}

		if action == security.ActionForceLogout {
			respondJSON(w, http.StatusUnauthorized, "यह खाता किसी अन्य डिवाइस पर खोला गया है।", security.ActionForceLogout)
			return
		}

		// सत्यापित यूज़र डेटा को रिक्वेस्ट कॉन्टेक्स्ट में आगे बढ़ाना
		ctx := context.WithValue(r.Context(), StudentUIDKey, studentUID)
		ctx = context.WithValue(ctx, DeviceIDKey, deviceID)
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

