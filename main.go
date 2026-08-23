package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// 1. रूट और एडमिन दोनों हैंडलर
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="hi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>अनंत अभ्यास एडमिन</title>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; background: #0f172a; color: #f8fafc; padding: 16px; margin: 0; }
        .card { background: #1e293b; padding: 16px; border-radius: 12px; margin-bottom: 16px; }
        .btn-wa { display: block; background: #25d366; color: white; text-align: center; padding: 14px; border-radius: 8px; font-weight: bold; text-decoration: none; margin-bottom: 16px; font-size: 16px; }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
        .metric { background: #0f172a; padding: 12px; border-radius: 8px; text-align: center; }
        .val { font-size: 22px; font-weight: bold; color: #38bdf8; margin-top: 4px; }
    </style>
</head>
<body>
    <h2>🎓 अनंत अभ्यास एडमिन पोर्टल</h2>
    <p style="color: #4ade80; margin-top: -8px;">गेटवे नंबर: +91 9664006651 (Live)</p>

    <!-- WhatsApp डायरेक्ट शेयर बटन -->
    <a class="btn-wa" href="https://wa.me/919664006651?text=राम%20राम%20सा%2C%20मुझे%20अनंत%20अभ्यास%20का%20फ्री%20डेमो%20चाहिए" target="_blank">
        📲 पेरेंट्स को WhatsApp लिंक भेजें
    </a>

    <div class="card">
        <h3>📊 लाइव मेट्रिक्स</h3>
        <div class="grid">
            <div class="metric"><small>कुल सक्रिय छात्र</small><div class="val">1</div></div>
            <div class="metric"><small>नए जुड़े बच्चे</small><div class="val" style="color:#4ade80;">+1</div></div>
            <div class="metric"><small>रेफरल पूरे हुए</small><div class="val" style="color:#a855f7;">0</div></div>
            <div class="metric"><small>पेंडिंग रेफरल</small><div class="val" style="color:#facc15;">0</div></div>
        </div>
    </div>

    <div class="card">
        <h3>🔍 360° छात्र कुंडली सर्च</h3>
        <input type="text" id="q" placeholder="UID या फोन नंबर दर्ज करें..." style="width: calc(100% - 24px); padding: 12px; background: #0f172a; border: 1px solid #475569; color: #fff; border-radius: 6px;">
        <button onclick="alert('छात्र: आरव (कक्षा 6) | 7-दिन फ्री डेमो सक्रिय')" style="margin-top: 10px; width: 100%; padding: 12px; background: #2563eb; color: #fff; border: none; border-radius: 6px; font-weight: bold; cursor: pointer;">सर्च करें</button>
    </div>
</body>
</html>`)
	})

	// 2. vCard 1-टैप कॉन्टैक्ट सेवर
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		vcard := strings.Join([]string{
			"BEGIN:VCARD",
			"VERSION:3.0",
			"N:मास्टरजी;अनंत अभ्यास;;;",
			"FN:अनंत अभ्यास - डिजिटल मास्टरजी",
			"ORG:Anant Abhyas Education;",
			"TEL;TYPE=CELL,VOICE,PREF:+919664006651",
			"NOTE:रोजाना 15 मिनट बोलकर अभ्यास और 7-दिन फ्री डेमो।",
			"URL:https://wa.me/919664006651?text=राम%20राम%20सा%20मुझे%20फ्री%20डेमो%20चाहिए",
			"END:VCARD",
		}, "\r\n")

		w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"Anant_Abhyas.vcf\"")
		w.Write([]byte(vcard))
	})

	// 3. पोर्ट बाइंडिंग
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	log.Printf("🚀 सर्वर पोर्ट %s पर सक्रिय है", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("सर्वर क्रैश: %v", err)
	}
}
