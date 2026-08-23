package admin

import (
	"database/sql"
	"net/http"
)

type AdminDashboard struct {
	db *sql.DB
}

func NewAdminDashboard(db *sql.DB) *AdminDashboard {
	return &AdminDashboard{db: db}
}

func (a *AdminDashboard) ServeDashboardHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="hi">
<head><meta charset="UTF-8"><title>अनंत अभ्यास - मास्टर एडमिन पैनल</title></head>
<body style="font-family:Segoe UI;background:#0f172a;color:#f8fafc;padding:20px;">
    <h2>🎓 अनंत अभ्यास: मास्टर कंट्रोल पैनल (गेटवे: 9664006651)</h2>
    <div style="background:#1e293b;padding:20px;border-radius:10px;margin-bottom:20px;">
        <h3>🔍 360° छात्र कुंडली सर्च</h3>
        <input type="text" id="searchKey" placeholder="UID या फोन..." style="padding:10px;width:60%;background:#0f172a;color:#fff;border:1px solid #475569;">
        <button onclick="searchKundali()" style="padding:10px 20px;background:#2563eb;color:#fff;border:none;cursor:pointer;">सर्च करें</button>
        <div id="resultBox" style="margin-top:15px;display:none;"></div>
    </div>
    <script>
        function searchKundali() {
            const k = document.getElementById('searchKey').value.trim();
            fetch('/api/v1/admin/kundali?q='+encodeURIComponent(k)).then(r=>r.json()).then(d=>{
                const b = document.getElementById('resultBox');
                b.style.display='block';
                b.innerHTML = '<p><strong>छात्र:</strong> '+d.student_name+' (कक्षा: '+d.class_level+') | <strong>भाषा:</strong> '+d.preferred_dialect+'</p>';
            });
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
