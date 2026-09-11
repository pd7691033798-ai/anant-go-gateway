package featurephone

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type CallerProfile struct {
	Name      string
	IsSpam    bool
	SpamScore int
	IsFriend  bool
}

type ContactFilter struct {
	mu          sync.RWMutex
	familyNodes map[string]string // phone -> name/label
	apiKey      string
	httpClient  *http.Client
}

func NewContactFilter() *ContactFilter {
	cf := &ContactFilter{
		familyNodes: make(map[string]string),
		apiKey:      os.Getenv("RAPIDAPI_TRUECALLER_KEY"),
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}

	// डिफ़ॉल्ट एडमिन नंबर
	cf.familyNodes["9024414973"] = "एडमिन (Self)"

	// whitelist.json फ़ाइल लोड करें
	_ = cf.LoadFromJSON("config/whitelist.json")
	return cf
}

// LoadFromJSON: JSON फ़ाइल से संपर्कों को सुरक्षित लोड करता है
func (cf *ContactFilter) LoadFromJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	cf.mu.Lock()
	defer cf.mu.Unlock()

	var contacts map[string]string
	if err := json.Unmarshal(data, &contacts); err != nil {
		return err
	}

	for phone, name := range contacts {
		clean := cleanNumber(phone)
		if clean != "" {
			cf.familyNodes[clean] = name
		}
	}
	return nil
}

// AddPersonalContact: रनटाइम पर लाइव नया पर्सनल नंबर जोड़ता है
func (cf *ContactFilter) AddPersonalContact(phone, label string) {
	clean := cleanNumber(phone)
	if clean == "" {
		return
	}

	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.familyNodes[clean] = label
}

// IsPersonalContact: केवल यह जांचता है कि नंबर पर्सनल लिस्ट में है या नहीं
func (cf *ContactFilter) IsPersonalContact(phone string) (bool, string) {
	clean := cleanNumber(phone)

	cf.mu.RLock()
	defer cf.mu.RUnlock()
	label, exists := cf.familyNodes[clean]
	return exists, label
}

// CheckCaller: पहले पर्सनल लिस्ट देखता है, अनजान नंबर होने पर Truecaller से स्कैन करता है
func (cf *ContactFilter) CheckCaller(rawPhone string) CallerProfile {
	clean := cleanNumber(rawPhone)

	cf.mu.RLock()
	name, exists := cf.familyNodes[clean]
	cf.mu.RUnlock()

	// 1. पर्सनल व्हाइटलिस्ट में मौजूद (दोस्त/परिवार/रिश्तेदार)
	if exists {
		return CallerProfile{
			Name:     name,
			IsFriend: true,
			IsSpam:   false,
		}
	}

	// 2. अनजान नंबर -> RapidAPI Truecaller से विवरण
	if cf.apiKey != "" {
		tcName, isSpam := cf.lookupTruecaller(clean)
		return CallerProfile{
			Name:     tcName,
			IsFriend: false,
			IsSpam:   isSpam,
		}
	}

	return CallerProfile{
		Name:     "Unknown Student",
		IsFriend: false,
		IsSpam:   false,
	}
}

func (cf *ContactFilter) lookupTruecaller(phone string) (string, bool) {
	reqURL := fmt.Sprintf("https://truecaller-data.p.rapidapi.com/search?phone=%s&countryCode=IN", url.QueryEscape(phone))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "New Student", false
	}

	req.Header.Add("X-RapidAPI-Key", cf.apiKey)
	req.Header.Add("X-RapidAPI-Host", "truecaller-data.p.rapidapi.com")

	resp, err := cf.httpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "New Student", false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var res struct {
		Data struct {
			Name      string `json:"name"`
			SpamScore int    `json:"spamScore"`
			Badges    []string `json:"badges"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &res); err == nil {
		name := strings.TrimSpace(res.Data.Name)
		if name == "" {
			name = "New Student"
		}
		isSpam := res.Data.SpamScore > 50
		return name, isSpam
	}

	return "New Student", false
}

func cleanNumber(val string) string {
	digits := ""
	for _, ch := range val {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}
	if len(digits) >= 10 {
		return digits[len(digits)-10:]
	}
	return digits
}

