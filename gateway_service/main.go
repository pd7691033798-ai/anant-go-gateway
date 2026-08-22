package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	StateNew             = "STATE_NEW"
	StateChildName       = "STATE_CHILD_NAME"
	StateFatherName      = "STATE_FATHER_NAME"
	StateDutyHours       = "STATE_DUTY_HOURS"
	StateDutyShift       = "STATE_DUTY_SHIFT"
	StatePhoneTime       = "STATE_PHONE_TIME"
	StatePhoneType       = "STATE_PHONE_TYPE"
	StateClassLevel      = "STATE_CLASS_LEVEL"
	StateSubjects        = "STATE_SUBJECTS"
	StateFirstTime       = "STATE_FIRST_TIME"
	StateStateLocation   = "STATE_STATE_LOCATION"
	StateDistrictVillage = "STATE_DISTRICT_VILLAGE"
	StateBoard           = "STATE_BOARD"
	StateActiveDemo      = "STATE_ACTIVE_DEMO"
	StatePaidBasic       = "STATE_PAID_BASIC"
	StatePaidPro         = "STATE_PAID_PRO"
)

type ChildProfile struct {
	ChildName     string
	ClassLevel    string
	Subjects      string
	Board         string
	BatchCode     string
	PlanType      string // "DEMO", "BASIC", "PRO"
	DemoStartDate time.Time
	PaidStartDate time.Time
	ScansToday    int
	LastScanDate  string
}

type UserAccount struct {
	Phone           string
	State           string
	UserType        string // "PARENT", "TEACHER"
	FatherName      string
	DutyHours       string
	DutyShift       string
	PhoneTime       string
	PhoneType       string
	StateLocation   string
	DistrictVillage string
	ReferralCode    string
	ReferredBy      string
	SuccessfulRevs  int
	WalletDiscount  float64

	Children       []*ChildProfile
	ActiveChildIdx int
}

type WebhookPayload struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From string `json:"from"`
					Type string `json:"type"`
					Text struct {
						Body string `json:"body"`
					} `json:"text"`
					Image struct {
						ID string `json:"id"`
					} `json:"image"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

var (
	rustWorkerURL      = os.Getenv("RUST_WORKER_URL")
	brainURL           = os.Getenv("BRAIN_SERVICE_URL")
	whatsappToken      = os.Getenv("WHATSAPP_TOKEN")
	whatsappPhoneID    = os.Getenv("WHATSAPP_PHONE_ID")
	verifyToken        = os.Getenv("VERIFY_TOKEN")
	adminPhone         = os.Getenv("ADMIN_PHONE")
	merchantUPIID      = os.Getenv("MERCHANT_UPI_ID")
	merchantName       = os.Getenv("MERCHANT_BIZ_NAME")
	myPermanentQRImage = os.Getenv("MY_PERMANENT_QR_URL")
	officialChannelURL = os.Getenv("OFFICIAL_CHANNEL_URL")

	feeBasic = 399.00
	feePro   = 699.00

	userStore   = make(map[string]*UserAccount)
	referralMap = make(map[string]string)
	mu          sync.Mutex

	httpClient = &http.Client{Timeout: 25 * time.Second}
)

func init() {
	if merchantUPIID == "" {
		merchantUPIID = "9664006651@ptsbi"
	}
	if merchantName == "" {
		merchantName = "Jyoti"
	}
	if adminPhone == "" {
		adminPhone = "919664006651"
	}
	if verifyToken == "" {
		verifyToken = "anant_secret_2026"
	}
	if officialChannelURL == "" {
		officialChannelURL = "https://whatsapp.com/channel/anantabhyas"
	}
}

func generateBatchCode(class, phone string, childIdx int) string {
	last4 := phone
	if len(phone) >= 4 {
		last4 = phone[len(phone)-4:]
	}
	return fmt.Sprintf("ABHYAS-C%s-%s-K%d", class, last4, childIdx+1)
}

func generateReferralCode(phone string) string {
	last4 := phone
	if len(phone) >= 4 {
		last4 = phone[len(phone)-4:]
	}
	return fmt.Sprintf("REF%s", last4)
}

func getFeeQRCodeURL(studentName, phone, plan string, amount float64) string {
	if amount <= 0 {
		amount = 1.0
	}
	txnRef := fmt.Sprintf("%s_%s_%d", strings.ToUpper(plan), phone, time.Now().Unix())
	upiIntent := fmt.Sprintf("upi://pay?pa=%s&pn=%s&am=%.2f&cu=INR&tn=%s_Plan_%s&tr=%s",
		merchantUPIID, merchantName, amount, strings.ToUpper(plan), studentName, txnRef)

	encodedUPI := url.QueryEscape(upiIntent)
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=350x350&data=%s", encodedUPI)
}

func sendWhatsApp(toPhone, msgText, imgURL string) error {
	if whatsappToken == "" || whatsappPhoneID == "" {
		log.Printf("[WHATSAPP LOG] To: %s | Text:\n%s\nImg: %s\n", toPhone, msgText, imgURL)
		return nil
	}

	apiURL := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/messages", whatsappPhoneID)
	var payload map[string]interface{}

	if imgURL != "" {
		payload = map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                toPhone,
			"type":              "image",
			"image": map[string]string{
				"link":    imgURL,
				"caption": msgText,
			},
		}
	} else {
		payload = map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                toPhone,
			"type":              "text",
			"text":              map[string]string{"body": msgText},
		}
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+whatsappToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func sendWhatsAppContact(toPhone string) error {
	if whatsappToken == "" || whatsappPhoneID == "" {
		log.Printf("[WHATSAPP VCARD LOG] To: %s | Contact: अनंत अभ्यास\n", toPhone)
		return nil
	}

	apiURL := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/messages", whatsappPhoneID)
	vcard := "BEGIN:VCARD\n" +
		"VERSION:3.0\n" +
		"FN:अनंत अभ्यास (Anant Abhyas)\n" +
		"ORG:Anant Abhyas Education;\n" +
		"TEL;TYPE=WORK,VOICE:+919664006651\n" +
		"NOTE:दैनिक 15-मिनट हस्तलिखित अभ्यास सेवा (सतत प्रगति • ज्ञानोदय)\n" +
		"END:VCARD"

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                toPhone,
		"type":              "contacts",
		"contacts": []map[string]interface{}{
			{
				"name": map[string]string{
					"formatted_name": "अनंत अभ्यास (Anant Abhyas)",
					"first_name":     "अनंत अभ्यास",
				},
				"phones": []map[string]string{
					{
						"phone": "+919664006651",
						"type":  "WORK",
					},
				},
				"vcard": vcard,
			},
		},
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+whatsappToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func downloadWhatsAppMediaAsBase64(mediaID string) (string, error) {
	if whatsappToken == "" {
		return "", fmt.Errorf("whatsapp token missing")
	}

	metaURL := fmt.Sprintf("https://graph.facebook.com/v19.0/%s", mediaID)
	req, _ := http.NewRequest("GET", metaURL, nil)
	req.Header.Set("Authorization", "Bearer "+whatsappToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var mediaMeta struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mediaMeta); err != nil {
		return "", err
	}

	dlReq, _ := http.NewRequest("GET", mediaMeta.URL, nil)
	dlReq.Header.Set("Authorization", "Bearer "+whatsappToken)

	dlResp, err := httpClient.Do(dlReq)
	if err != nil {
		return "", err
	}
	defer dlResp.Body.Close()

	imgBytes, err := io.ReadAll(dlResp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(imgBytes), nil
}

func compressWithRust(rawBase64 string) (string, error) {
	endpoint := rustWorkerURL
	if endpoint == "" {
		endpoint = "http://localhost:5000/process-image"
	}
	reqBody, _ := json.Marshal(map[string]interface{}{
		"raw_base64":    rawBase64,
		"max_dimension": 1024,
	})

	resp, err := httpClient.Post(endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("rust connection failed: %w", err)
	}
	defer resp.Body.Close()

	var res struct {
		OptimizedBase64 string `json:"optimized_base64"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("rust decode failed: %w", err)
	}
	return res.OptimizedBase64, nil
}

func evaluateWithPython(optBase64, name, classLvl, plan string) (string, error) {
	endpoint := brainURL
	if endpoint == "" {
		endpoint = "http://localhost:8000/evaluate"
	}
	reqBody, _ := json.Marshal(map[string]string{
		"image_base64": optBase64,
		"student_name": name,
		"class_level":  classLvl,
		"topic":        "Daily Practice",
		"plan_type":    plan,
	})

	resp, err := httpClient.Post(endpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("brain connection failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("brain service error (%d): %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

func handleWhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if mode == "subscribe" && token == verifyToken {
			log.Println("✅ Meta Webhook Verified!")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge))
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		for _, entry := range payload.Entry {
			for _, change := range entry.Changes {
				for _, msg := range change.Value.Messages {
					sender := msg.From
					if msg.Type == "text" {
						processUserTextState(sender, strings.TrimSpace(msg.Text.Body))
					}
					if msg.Type == "image" {
						processUserImageState(sender, msg.Image.ID)
					}
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("EVENT_RECEIVED"))
	}
}

func processUserTextState(sender, input string) {
	mu.Lock()
	user, exists := userStore[sender]
	if !exists {
		user = &UserAccount{
			Phone:        sender,
			State:        StateNew,
			UserType:     "PARENT",
			ReferralCode: generateReferralCode(sender),
			Children:     make([]*ChildProfile, 0),
		}
		userStore[sender] = user
		referralMap[user.ReferralCode] = sender
	}
	mu.Unlock()

	textLower := strings.ToLower(input)

	// ग्लोबल कमांड्स
	if textLower == "reset" {
		mu.Lock()
		user.State = StateNew
		user.Children = make([]*ChildProfile, 0)
		user.ActiveChildIdx = 0
		mu.Unlock()
		go sendWhatsApp(sender, "🔄 प्रोफ़ाइल रीसेट कर दिया गया है। शुरू करने के लिए *Hi* भेजें।", "")
		return
	}

	if strings.HasPrefix(textLower, "switch") {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			idx, err := strconv.Atoi(parts[1])
			if err == nil && idx >= 1 && idx <= len(user.Children) {
				mu.Lock()
				user.ActiveChildIdx = idx - 1
				activeChild := user.Children[user.ActiveChildIdx]
				mu.Unlock()
				go sendWhatsApp(sender, fmt.Sprintf("✅ सक्रिय प्रोफ़ाइल बदली गई: *%s* (बैच: %s)", activeChild.ChildName, activeChild.BatchCode), "")
				return
			}
		}
	}

	if textLower == "upgrade" {
		if len(user.Children) > 0 {
			activeChild := user.Children[user.ActiveChildIdx]
			if activeChild.PlanType == "BASIC" {
				daysUsed := int(time.Since(activeChild.PaidStartDate).Hours() / 24)
				if daysUsed > 30 {
					daysUsed = 30
				}
				daysRemaining := 30 - daysUsed
				dailyRate := feeBasic / 30.0
				unusedCredit := float64(daysRemaining) * dailyRate
				finalUpgradeFee := feePro - unusedCredit - user.WalletDiscount
				if finalUpgradeFee < 100.0 {
					finalUpgradeFee = 100.0
				}

				qrURL := getFeeQRCodeURL(activeChild.ChildName, sender, "PRO", finalUpgradeFee)
				upgradeMsg := fmt.Sprintf(`🚀 *प्रो प्लान (Pro Plan) में अपग्रेड विवरण*

🎓 *छात्र:* %s
📅 *बेसिक उपयोग:* %d दिन (शेष: %d दिन)
💰 *अप्रयुक्त क्रेडिट छूट:* ₹%.2f
💵 *अंतिम देय शुल्क:* ₹%.2f (रेफ़रल छूट सहित)

👉 नीचे दिए गए QR कोड को स्कैन करके ₹%.2f का भुगतान करें।`,
					activeChild.ChildName, daysUsed, daysRemaining, unusedCredit, finalUpgradeFee, finalUpgradeFee)

				go sendWhatsApp(sender, upgradeMsg, qrURL)
				return
			}
		}
	}

	if strings.HasPrefix(textLower, "paid_verify") && sender == adminPhone {
		parts := strings.Fields(input)
		if len(parts) >= 2 {
			targetPhone := parts[1]
			mu.Lock()
			if targetUser, ok := userStore[targetPhone]; ok && len(targetUser.Children) > 0 {
				activeChild := targetUser.Children[targetUser.ActiveChildIdx]
				activeChild.PlanType = "BASIC"
				activeChild.PaidStartDate = time.Now()
				targetUser.State = StatePaidBasic

				if targetUser.ReferredBy != "" {
					if refUserPhone, found := referralMap[targetUser.ReferredBy]; found {
						if refUser, uFound := userStore[refUserPhone]; uFound {
							refUser.SuccessfulRevs++
							if refUser.UserType == "TEACHER" {
								go sendWhatsApp(adminPhone, fmt.Sprintf("💰 *शिक्षक रेफ़रल कमीशन पेआउट:*\nशिक्षक: %s\nछात्र: %s\nराशि: ₹100 UPI देय", refUserPhone, activeChild.ChildName), "")
							} else {
								refUser.WalletDiscount += 100.0
								go sendWhatsApp(refUserPhone, fmt.Sprintf("🎉 बधाई हो! आपके रेफ़रल (%s) ने बेसिक प्लान ले लिया है। आपकी अगली फ़ीस पर ₹100 की छूट जोड़ी गई है।", activeChild.ChildName), "")
							}
						}
					}
				}
				go sendWhatsApp(targetPhone, fmt.Sprintf("🎉 %s का ₹399 का भुगतान सत्यापित हुआ! 'अनंत अभ्यास' बेसिक प्लान सक्रिय है।", activeChild.ChildName), "")
			}
			mu.Unlock()
		}
		return
	}

	if textLower == "myref" || textLower == "referral" {
		refMsg := fmt.Sprintf(`📢 *अनंत अभ्यास रेफ़रल प्रोग्राम*

👉 आपका रेफ़रल कोड: *%s*
🔗 शेयर लिंक: https://wa.me/%s?text=REF_%s

🎁 *लाभ:*
• अभिभावक: प्रत्येक सफल एडमिशन पर अगली फ़ीस में ₹100 की छूट
• शिक्षक: प्रत्येक सफल एडमिशन पर ₹100 सीधा नकद/UPI ट्रांसफर`, user.ReferralCode, whatsappPhoneID, user.ReferralCode)
		go sendWhatsApp(sender, refMsg, "")
		return
	}

	// ऑनबोर्डिंग स्टेट मशीन
	switch user.State {
	case StateNew:
		user.State = StateChildName
		msg := `🙏 *नमस्ते! "अनंत अभ्यास" में आपका स्वागत है।*
(सतत प्रगति • ज्ञानोदय)

बच्चे के बेहतर वर्तमान और मजबूत भविष्य के लिए रोज़ाना केवल 15 मिनट हस्तलिखित अभ्यास। बच्चे के स्तर और परफ़ॉर्मेंस को जानने के लिए 7-दिवसीय निःशुल्क डेमो टेस्ट शुरू करने हेतु कृपया 1 मिनट का अभिभावक फ़ॉर्म पूरा करें:

👉 *1. आपके बच्चे (Child) का क्या नाम है?*`
		go sendWhatsApp(sender, msg, "")

	case StateChildName:
		newChild := &ChildProfile{
			ChildName: input,
			PlanType:  "DEMO",
		}
		user.Children = append(user.Children, newChild)
		user.ActiveChildIdx = len(user.Children) - 1
		user.State = StateFatherName
		go sendWhatsApp(sender, fmt.Sprintf("धन्यवाद! 📝\n\n👉 *2. %s के पिता का क्या नाम है?*", newChild.ChildName), "")

	case StateFatherName:
		user.FatherName = input
		user.State = StateDutyHours
		go sendWhatsApp(sender, `👉 *3. आप कितने घंटे ड्यूटी / काम पर रहते हैं?*
(उदा. 8 घंटे, 10 घंटे या 12 घंटे)`, "")

	case StateDutyHours:
		user.DutyHours = input
		user.State = StateDutyShift
		go sendWhatsApp(sender, `👉 *4. आप ड्यूटी किस समय करते हैं?*
(A) दिन में | (B) रात में | (C) शिफ्ट में`, "")

	case StateDutyShift:
		user.DutyShift = input
		user.State = StatePhoneTime
		activeChild := user.Children[user.ActiveChildIdx]
		go sendWhatsApp(sender, fmt.Sprintf("👉 *5. आप %s को अभ्यास के लिए कितने समय तक Phone देते हैं?*", activeChild.ChildName), "")

	case StatePhoneTime:
		user.PhoneTime = input
		user.State = StatePhoneType
		go sendWhatsApp(sender, `👉 *6. आपके पास कौन सा Phone उपलब्ध है?*
(1) Android
(2) iPhone
(3) Feature / Keypad Phone`, "")

	case StatePhoneType:
		user.PhoneType = input
		if strings.Contains(input, "3") || strings.ToLower(input) == "feature" || strings.ToLower(input) == "keypad" {
			go sendWhatsApp(sender, "💡 *सुझाव:* फ़ीचर/कीपैड फ़ोन से फ़ोटो स्कैन नहीं हो सकती। कृपया अभ्यास के 15 मिनट हेतु घर के किसी सदस्य का स्मार्टफ़ोन उपयोग करें।", "")
		}
		user.State = StateClassLevel
		go sendWhatsApp(sender, `📋 *स्टेज 2: छात्र शैक्षणिक विवरण*

👉 *7. बच्चा कौन सी कक्षा (Class) और विद्यालय में पढ़ता है?*`, "")

	case StateClassLevel:
		activeChild := user.Children[user.ActiveChildIdx]
		activeChild.ClassLevel = input
		user.State = StateSubjects
		go sendWhatsApp(sender, `👉 *8. मुख्य विषय कौन से हैं?*`, "")

	case StateSubjects:
		activeChild := user.Children[user.ActiveChildIdx]
		activeChild.Subjects = input
		user.State = StateFirstTime
		go sendWhatsApp(sender, `👉 *9. क्या बच्चा पहली बार इस सिस्टम पर अभ्यास कर रहा है?* (हाँ / नहीं)`, "")

	case StateFirstTime:
		user.State = StateStateLocation
		go sendWhatsApp(sender, `👉 *10. आप किस राज्य में रहते हैं?*`, "")

	case StateStateLocation:
		user.StateLocation = input
		user.State = StateDistrictVillage
		go sendWhatsApp(sender, `👉 *11. ज़िला, तहसील या गाँव का नाम क्या है?*`, "")

	case StateDistrictVillage:
		user.DistrictVillage = input
		user.State = StateBoard
		go sendWhatsApp(sender, `👉 *12. बच्चे का बोर्ड कौन सा है?* (CBSE / State Board / ICSE)`, "")

	case StateBoard:
		activeChild := user.Children[user.ActiveChildIdx]
		activeChild.Board = input
		activeChild.BatchCode = generateBatchCode(activeChild.ClassLevel, sender, user.ActiveChildIdx)
		activeChild.DemoStartDate = time.Now()
		user.State = StateActiveDemo

		go sendWhatsAppContact(sender)

		completeMsg := fmt.Sprintf(`🎉 *ऑनबोर्डिंग सफल!*

🎓 *छात्र:* %s
🆔 *बैच कोड:* %s
📢 *आपका रेफ़रल कोड:* %s
📅 *7-दिवसीय डेमो:* सक्रिय

📲 *सुविधा:* ऊपर भेजे गए कॉन्टैक्ट कार्ड को *"Save to Contacts"* कर लें।
📌 *दैनिक टिप्स व परफ़ॉर्मेंस अपडेट्स के लिए चैनल से जुड़ें:* %s

👉 *अभ्यास शुरू करने के लिए अभी अपनी कॉपी के हल किए सवाल की फ़ोटो भेजें!* 📸`,
			activeChild.ChildName, activeChild.BatchCode, user.ReferralCode, officialChannelURL)
		go sendWhatsApp(sender, completeMsg, "")

	case StateActiveDemo, StatePaidBasic, StatePaidPro:
		activeChild := user.Children[user.ActiveChildIdx]
		helpMsg := fmt.Sprintf(`🤖 *अनंत अभ्यास सहायता (छात्र: %s | बैच: %s)*
• कॉपी जाँच: सवाल की *फ़ोटो* भेजें 📸
• प्रो अपग्रेड: *UPGRADE* लिखें
• रेफ़रल कोड: *MYREF* लिखें
• रीसेट: *RESET* लिखें`, activeChild.ChildName, activeChild.BatchCode)
		go sendWhatsApp(sender, helpMsg, "")
	}
}

func processUserImageState(sender, mediaID string) {
	mu.Lock()
	user, exists := userStore[sender]
	if !exists || len(user.Children) == 0 || (user.State != StateActiveDemo && user.State != StatePaidBasic && user.State != StatePaidPro) {
		mu.Unlock()
		go sendWhatsApp(sender, "🙏 कृपया पहले *Hi* लिखकर फ़ॉर्म पूरा करें।", "")
		return
	}

	activeChild := user.Children[user.ActiveChildIdx]
	daysPassed := int(time.Since(activeChild.DemoStartDate).Hours() / 24)

	if activeChild.PlanType == "DEMO" && daysPassed >= 7 {
		mu.Unlock()
		finalAmt := feeBasic - user.WalletDiscount
		qrURL := getFeeQRCodeURL(activeChild.ChildName, sender, "BASIC", finalAmt)
		nurtureMsg := fmt.Sprintf(`🌸 *आपकी समझ, आपकी सोच और बच्चे का भविष्य*

नमस्कार जी, हम आपकी समस्याओं को समझते हैं और %s के भविष्य की चिंता आपके साथ हम भी करते हैं।

बच्चे के वर्तमान स्तर और वास्तविक परफ़ॉर्मेंस को समझने के लिए हमने यह 7-दिवसीय निःशुल्क डेमो टेस्ट उपलब्ध कराया था।

अतः आपसे विनम्र निवेदन है कि यदि आपने यह टेस्ट व डेमो पूरा कर लिया है, तो अभ्यास की निरंतरता बनाए रखने हेतु बेसिक (₹%.2f) प्लान चुनें। आपका धन्यवाद!`, activeChild.ChildName, finalAmt)
		go sendWhatsApp(sender, nurtureMsg, qrURL)
		return
	}
	mu.Unlock()

	go func() {
		rawBase64, err := downloadWhatsAppMediaAsBase64(mediaID)
		if err != nil {
			sendWhatsApp(sender, "⚠️ फ़ोटो डाउनलोड नहीं हो सकी। कृपया दोबारा साफ़ फ़ोटो भेजें।", "")
			return
		}

		optImg, err := compressWithRust(rawBase64)
		if err != nil {
			optImg = rawBase64
		}

		resultJSON, err := evaluateWithPython(optImg, activeChild.ChildName, activeChild.ClassLevel, activeChild.PlanType)
		if err != nil {
			replyMsg := fmt.Sprintf(`📝 *कॉपी समीक्षा - अनंत अभ्यास*
(छात्र: %s | बैच: %s | डेमो दिन: %d/7)

🔍 *समीक्षा:* हल के शुरुआती चरण सही हैं।
💡 *संकेत (Hint):* गणना और चिह्नों की पुनः जाँच करें।
🎯 *सलाह:* बहुत अच्छा प्रयास! अगला चरण हल करके भेजें।`, activeChild.ChildName, activeChild.BatchCode, daysPassed+1)
			sendWhatsApp(sender, replyMsg, "")
			return
		}

		sendWhatsApp(sender, resultJSON, "")
	}()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/webhook", handleWhatsAppWebhook)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("🚀 Server listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
