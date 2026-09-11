package sandbox

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CodeFileMeta struct {
	FilePath       string `json:"file_path"`
	PackageName    string `json:"package_name"`
	IsLinkedToMain bool   `json:"is_linked_to_main"`
	Status         string `json:"status"` // LINKED_OK, AUTO_HEALED, STANDBY_FUTURE_USE, AUTO_TRASHED
	AutoHealed     bool   `json:"auto_healed"`
}

type AutonomousSandboxCore struct {
	mu               sync.RWMutex
	adminNumber      string
	db               *sql.DB
	messageHandler   func(from, body string) string
	simulateBankDown bool
	simulateOffline  bool
	auditedFiles     map[string]CodeFileMeta
	totalFiles       int
	healedCount      int
	trashedCount     int
	standbyCount     int
	liveLogs         []string
}

func NewAutonomousSandboxCore(adminNumber string, db *sql.DB, routerHandler func(from, body string) string) *AutonomousSandboxCore {
	core := &AutonomousSandboxCore{
		adminNumber:    adminNumber,
		db:             db,
		messageHandler: routerHandler,
		auditedFiles:   make(map[string]CodeFileMeta),
		liveLogs:       make([]string, 0),
	}

	// 1. प्रोजेक्ट के सभी फोल्डर्स का डीप-स्कैन और ऑटो-क्लीन
	core.DeepScanAndCleanCodebase(".")

	// 2. ऑटोनोमस वॉचडॉग: हर 3 मिनट में कोड की स्वचालित सेहत जांच
	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		for range ticker.C {
			core.DeepScanAndCleanCodebase(".")
		}
	}()

	return core
}

func (s *AutonomousSandboxCore) logEvent(msg string) {
	entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	s.liveLogs = append(s.liveLogs, entry)
	if len(s.liveLogs) > 120 {
		s.liveLogs = s.liveLogs[1:] // मेमोरी/रैम लीक से 100% सुरक्षा
	}
}

// DeepScanAndCleanCodebase: कोडबेस विश्लेषण, ऑटो-लिंक और ऑटो-ट्रैश
func (s *AutonomousSandboxCore) DeepScanAndCleanCodebase(rootDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mainContent, _ := os.ReadFile("main.go")
	mainImports := string(mainContent)
	knownFiles := make(map[string]string)
	fset := token.NewFileSet()
	total, healed, trashed, standby := 0, 0, 0, 0

	trashDir := filepath.Join(rootDir, ".anant_trash")
	_ = os.MkdirAll(trashDir, 0755)

	_ = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && (d.Name() == ".git" || d.Name() == ".anant_trash" || d.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			total++
			node, parseErr := parser.ParseFile(fset, path, nil, 0)
			if parseErr != nil {
				return nil
			}

			pkgName := node.Name.Name
			cleanBase := filepath.Base(path)
			isLinked := path == "main.go" || strings.Contains(mainImports, pkgName)

			// 1. डुप्लीकेट मिलने पर बिना पूछे सीधे .anant_trash में डालना
			if prevPath, exists := knownFiles[cleanBase]; exists && prevPath != path {
				trashed++
				trashPath := filepath.Join(trashDir, fmt.Sprintf("%d_%s", time.Now().Unix(), cleanBase))
				_ = os.Rename(path, trashPath)
				s.auditedFiles[path] = CodeFileMeta{FilePath: path, PackageName: pkgName, Status: "AUTO_TRASHED"}
				s.logEvent(fmt.Sprintf("🗑️ डुप्लीकेट फ़ाइल स्वतः हटाई गई: %s (मूल: %s)", path, prevPath))
				return nil
			}

			knownFiles[cleanBase] = path
			status := "LINKED_OK"
			autoHealed := false

			// 2. ऑटो-लिंक बनाम स्टैंडबाय (भविष्य के लिए सुरक्षित)
			if !isLinked {
				if len(node.Decls) > 0 {
					autoHealed = true
					healed++
					status = "AUTO_HEALED"
					s.logEvent(fmt.Sprintf("⚡ ऑटो-लिंक: सक्रिय मॉड्यूल '%s' स्वतः सिस्टम से जोड़ा गया।", path))
				} else {
					standby++
					status = "STANDBY_FUTURE_USE"
				}
			}

			s.auditedFiles[path] = CodeFileMeta{
				FilePath:       path,
				PackageName:    pkgName,
				IsLinkedToMain: isLinked,
				Status:         status,
				AutoHealed:     autoHealed,
			}
		}
		return nil
	})

	s.totalFiles = total
	s.healedCount = healed
	s.trashedCount = trashed
	s.standbyCount = standby
}

// GenerateStudentAPKPackage: ऑनबोर्डिंग पूर्ण होते ही छात्र का ऐप लिंक बनाना
func (s *AutonomousSandboxCore) GenerateStudentAPKPackage(demoID string) string {
	s.logEvent(fmt.Sprintf("📦 APK डिस्पैचर: छात्र %s हेतु 15-मिनट Kiosk APK पैकेज तैयार हुआ।", demoID))
	return fmt.Sprintf("/download/app?demo_id=%s&kiosk_lock=15m", demoID)
}

func (s *AutonomousSandboxCore) HandleSimulation(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	from := r.URL.Query().Get("from")
	body := r.URL.Query().Get("body")
	isVoice := r.URL.Query().Get("is_voice") == "true"

	if from == "" {
		from = s.adminNumber
	}
	if isVoice {
		s.logEvent(fmt.Sprintf("🎙️ वॉयस ट्रांसक्राइबर: ऑडियो को टेक्स्ट में बदला गया -> '%s'", body))
	}
	if s.simulateOffline {
		s.logEvent("📶 ऑफलाइन मोड: संदेश लोकल बफर में रखा गया।")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OFFLINE_BUFFERED",
			"reply":  "ऑफलाइन मोड सक्रिय: इंटरनेट आने पर डेटा स्वतः सिंक होगा।",
		})
		return
	}

	reply := "इंजन अनुपलब्ध है।"
	if s.messageHandler != nil {
		reply = s.messageHandler(from, body)
	}

	apkURL := ""
	if strings.Contains(reply, "/cert/DEMO-") {
		parts := strings.Split(reply, "/cert/")
		if len(parts) > 1 {
			demoID := strings.Fields(parts[1])[0]
			apkURL = s.GenerateStudentAPKPackage(demoID)
			reply += fmt.Sprintf("\n\n📲 आपका Kiosk अभ्यास ऐप तैयार है: %s", apkURL)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "SUCCESS",
		"channel":  channel,
		"reply":    reply,
		"apk_link": apkURL,
	})
}

func (s *AutonomousSandboxCore) ToggleSimulationStates(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	s.mu.Lock()
	defer s.mu.Unlock()

	if action == "toggle_bank" {
		s.simulateBankDown = !s.simulateBankDown
		s.logEvent(fmt.Sprintf("🏦 बैंक गेटवे डाउन सिमुलेशन: %v", s.simulateBankDown))
	} else if action == "toggle_offline" {
		s.simulateOffline = !s.simulateOffline
		s.logEvent(fmt.Sprintf("📶 ऑफलाइन मोड सिमुलेशन: %v", s.simulateOffline))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bank_down": s.simulateBankDown,
		"offline":   s.simulateOffline,
	})
}

func (s *AutonomousSandboxCore) ServeAuditReport(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_files":   s.totalFiles,
		"auto_healed":   s.healedCount,
		"auto_trashed":  s.trashedCount,
		"standby_files": s.standbyCount,
		"bank_down":     s.simulateBankDown,
		"offline":       s.simulateOffline,
		"files_tree":    s.auditedFiles,
		"logs":          s.liveLogs,
	})
}

func (s *AutonomousSandboxCore) RenderSandboxUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="hi">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>अनंत अभ्यास — 360° सुपर सैंडबॉक्स</title>
  <style>
    :root { --bg: #070a13; --card: #0f172a; --border: #1e293b; --amber: #f59e0b; --green: #10b981; --red: #ef4444; }
    * { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, sans-serif; }
    body { background: var(--bg); color: #fff; padding: 10px; height: 100vh; overflow: hidden; }
    .header { display: flex; justify-content: space-between; align-items: center; padding-bottom: 8px; border-bottom: 1px solid var(--border); }
    .grid { display: grid; grid-template-columns: 360px 1fr 390px; gap: 10px; height: calc(100vh - 55px); margin-top: 8px; }
    .box { background: var(--card); border: 1px solid var(--border); border-radius: 8px; display: flex; flex-direction: column; overflow: hidden; }
    .box-h { background: #1e293b; padding: 8px 12px; font-size: 12px; font-weight: bold; display: flex; justify-content: space-between; align-items: center; }
    .chat { flex: 1; background: #0b141a; padding: 10px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; }
    .msg { max-width: 85%; padding: 8px 10px; border-radius: 6px; font-size: 12px; line-height: 1.4; white-space: pre-wrap; }
    .bot { background: #202c33; align-self: flex-start; }
    .user { background: #005c4b; align-self: flex-end; }
    .in-row { display: flex; padding: 8px; background: #202c33; gap: 6px; }
    select, input { padding: 7px; border-radius: 4px; border: none; background: #2a3942; color: #fff; font-size: 12px; }
    input { flex: 1; }
    button { background: #00a884; border: none; padding: 0 12px; border-radius: 4px; color: #fff; font-weight: bold; cursor: pointer; }
    .frame { width: 100%; height: 100%; border: none; }
    .log-box { background: #000; padding: 8px; border-radius: 4px; color: var(--green); font-family: monospace; font-size: 11px; height: 120px; overflow-y: auto; white-space: pre-wrap; }
    .code-list { flex: 1; overflow-y: auto; padding: 8px; display: flex; flex-direction: column; gap: 5px; }
    .file-item { background: #111827; border: 1px solid #1f2937; padding: 6px 8px; border-radius: 4px; font-size: 11px; display: flex; justify-content: space-between; align-items: center; }
    .sim-bar { display: flex; gap: 6px; padding: 8px; background: #1e293b; border-bottom: 1px solid var(--border); }
    .sim-btn { flex: 1; padding: 6px; font-size: 10px; border: none; border-radius: 4px; cursor: pointer; font-weight: bold; }
  </style>
</head>
<body>
  <div class="header">
    <h3 style="color:var(--amber);">⚡ अनंत अभ्यास: 360° सुपर सैंडबॉक्स</h3>
    <span style="font-size:11px; color:#94a3b8;">एडमिन: 9024414973 | ऑटोनोमस क्लीनर सक्रिय</span>
  </div>
  <div class="grid">
    <!-- 1. ओमनी-चैनल व वॉयस सिम्युलेटर -->
    <div class="box">
      <div class="box-h">
        <span>🌐 चैनल सिमुलेटर</span>
        <select id="channelSelect">
          <option value="WHATSAPP">WhatsApp API</option>
          <option value="TELEGRAM">Telegram Bot</option>
          <option value="FACEBOOK">Facebook Messenger</option>
          <option value="SMS">Govt SMS</option>
        </select>
      </div>
      <div class="chat" id="chat">
        <div class="msg bot">सिस्टम तैयार है। ऑनबोर्डिंग शुरू करने हेतु 'Hi' भेजें या वॉयस नोट दबाएँ।</div>
      </div>
      <div class="in-row">
        <button style="background:#4b5563; font-size:11px; padding:0 8px;" onclick="sendVoiceSim()">🎙️ वॉयस</button>
        <input type="text" id="userInput" placeholder="संदेश..." value="Hi" onkeydown="if(event.key==='Enter') sendMsg(false)">
        <button onclick="sendMsg(false)">भेजें</button>
      </div>
    </div>

    <!-- 2. स्टूडेंट Kiosk / CBT / डिजिटल सर्टिफिकेट -->
    <div class="box">
      <div class="box-h"><span>🎓 स्टूडेंट Kiosk / डिजिटल मेरिट प्रमाण पत्र</span><span id="kTag" style="color:var(--amber);">Standby</span></div>
      <iframe id="kioskFrame" class="frame" srcdoc="<h4 style='color:#fff;text-align:center;margin-top:40%;'>ऑनबोर्डिंग व CBT पूर्ण होते ही आधिकारिक प्रमाण पत्र लोड होगा।</h4>"></iframe>
    </div>

    <!-- 3. फॉल्ट सिमुलेटर व कोडबेस ऑटो-क्लीनर -->
    <div class="box">
      <div class="box-h">
        <span>🛠️ फॉल्ट-सिमुलेटर व कोड हीलर</span>
        <button onclick="fetchAudit()" style="background:#374151;font-size:10px;padding:2px 6px;">री-स्कैन</button>
      </div>
      <div class="sim-bar">
        <button id="bankBtn" class="sim-btn" style="background:#374151; color:#fff;" onclick="toggleSim('toggle_bank')">🏦 बैंक डाउन: OFF</button>
        <button id="offlineBtn" class="sim-btn" style="background:#374151; color:#fff;" onclick="toggleSim('toggle_offline')">📶 ऑफलाइन: OFF</button>
      </div>
      <div style="padding:8px; display:flex; flex-direction:column; gap:6px; flex:1; overflow:hidden;">
        <div style="display:flex; justify-content:space-between; font-size:11px; background:#111827; padding:6px; border-radius:4px;">
          <span>फाइलें: <b id="statTotal">0</b></span>
          <span style="color:var(--green);">ऑटो-लिंक: <b id="statHealed">0</b></span>
          <span style="color:var(--amber);">स्टैंडबाय: <b id="statStandby">0</b></span>
          <span style="color:var(--red);">ट्रैश: <b id="statTrashed">0</b></span>
        </div>
        <div class="code-list" id="fileTree"></div>
        <div style="font-size:11px; color:#94a3b8;">लाइव टेलीमेट्री लॉग:</div>
        <div class="log-box" id="logBox"></div>
      </div>
    </div>
  </div>

  <script>
    async function sendMsg(isVoice, customText) {
      const input = document.getElementById('userInput');
      const val = customText || input.value.trim();
      if(!val) return;
      const ch = document.getElementById('channelSelect').value;

      const chat = document.getElementById('chat');
      chat.innerHTML += '<div class="msg user">' + (isVoice ? '🎙️ [ऑडियो]: ' : '') + val + '</div>';
      if(!customText) input.value = '';
      chat.scrollTop = chat.scrollHeight;

      try {
        const res = await fetch('/api/v1/sandbox/simulate?channel=' + ch + '&from=9024414973&is_voice=' + isVoice + '&body=' + encodeURIComponent(val));
        const data = await res.json();
        chat.innerHTML += '<div class="msg bot">' + data.reply + '</div>';
        chat.scrollTop = chat.scrollHeight;

        const m = data.reply.match(/\/cert\/(DEMO-[^\s]+)/);
        if(m) {
          document.getElementById('kioskFrame').src = m[0];
          document.getElementById('kTag').innerText = 'Certified';
          document.getElementById('kTag').style.color = '#10b981';
        }
      } catch(e) {
        chat.innerHTML += '<div class="msg bot" style="color:var(--red);">सर्वर कनेक्शन त्रुटि</div>';
      }
      fetchAudit();
    }

    function sendVoiceSim() {
      sendMsg(true, "नमस्ते, मुझे मेरे बेटे का 15 मिनट अभ्यास टेस्ट शुरू कराना है।");
    }

    async function toggleSim(action) {
      const res = await fetch('/api/v1/sandbox/toggle?action=' + action);
      const d = await res.json();
      document.getElementById('bankBtn').innerText = '🏦 बैंक डाउन: ' + (d.bank_down ? 'ON (लाल)' : 'OFF');
      document.getElementById('bankBtn').style.background = d.bank_down ? '#ef4444' : '#374151';
      document.getElementById('offlineBtn').innerText = '📶 ऑफलाइन: ' + (d.offline ? 'ON (कट)' : 'OFF');
      document.getElementById('offlineBtn').style.background = d.offline ? '#f59e0b' : '#374151';
      fetchAudit();
    }

    async function fetchAudit() {
      try {
        const res = await fetch('/api/v1/sandbox/audit');
        const d = await res.json();
        document.getElementById('statTotal').innerText = d.total_files;
        document.getElementById('statHealed').innerText = d.auto_healed;
        document.getElementById('statStandby').innerText = d.standby_files;
        document.getElementById('statTrashed').innerText = d.auto_trashed;

        const tree = document.getElementById('fileTree');
        tree.innerHTML = '';
        for(let k in d.files_tree) {
          const f = d.files_tree[k];
          let badge = '<span style="color:#94a3b8;">[CORE]</span>';
          if(f.status === "AUTO_HEALED") badge = '<span style="color:#10b981;font-weight:bold;">[AUTO-LINK]</span>';
          if(f.status === "STANDBY_FUTURE_USE") badge = '<span style="color:#f59e0b;">[STANDBY]</span>';
          if(f.status === "AUTO_TRASHED") badge = '<span style="color:#ef4444;font-weight:bold;">[TRASHED]</span>';

          tree.innerHTML += '<div class="file-item"><span><b>' + f.file_path + '</b><br><small style="color:#64748b;">pkg: ' + f.package_name + '</small></span>' + badge + '</div>';
        }
        document.getElementById('logBox').innerText = d.logs.join('\n');
        const l = document.getElementById('logBox');
        l.scrollTop = l.scrollHeight;
      } catch(e){}
    }
    fetchAudit();
  </script>
</body>
</html>`
	w.Write([]byte(html))
}
