import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const String kBackendBaseUrl = "http://localhost:8080";

void main() {
  runApp(const MaterialApp(
    debugShowCheckedModeBanner: false,
    home: AdminUnifiedDashboard(),
  ));
}

class AdminUnifiedDashboard extends StatefulWidget {
  const AdminUnifiedDashboard({super.key});

  @override
  State<AdminUnifiedDashboard> createState() => _AdminUnifiedDashboardState();
}

class _AdminUnifiedDashboardState extends State<AdminUnifiedDashboard> with SingleTickerProviderStateMixin {
  late TabController _tabController;
  Map<String, dynamic> _clusterStats = {};

  final List<Map<String, dynamic>> _students = [
    {
      "name": "आर्यन शर्मा",
      "demoID": "DEMO-2026-4973",
      "claimed": 5,
      "detected": 8,
      "speed": "4s",
      "score": "16/20",
      "flagged": true,
      "wing": "नवोदय / सैनिक विंग"
    },
    {
      "name": "रोहन वर्मा",
      "demoID": "DEMO-2026-6651",
      "claimed": 11,
      "detected": 11,
      "speed": "29s",
      "score": "18/20",
      "flagged": false,
      "wing": "JEE Mains & Adv CBT"
    }
  ];

  final List<Map<String, dynamic>> _partnerLedger = [
    {
      "shop": "रमेश बुक डिपो (अलवर)",
      "code": "SHOP_RAM_4149",
      "conversions": 6,
      "payout": 600.0,
      "upi": "ramesh@oksbi",
      "isSettled": false
    },
    {
      "shop": "बालाजी कैफे (जयपुर)",
      "code": "SHOP_BAL_6651",
      "conversions": 4,
      "payout": 400.0,
      "upi": "balaji@paytm",
      "isSettled": true
    }
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _fetchClusterStats();
  }

  Future<void> _fetchClusterStats() async {
    try {
      final res = await http.get(Uri.parse('$kBackendBaseUrl/api/v1/admin/stats'));
      if (res.statusCode == 200) {
        setState(() => _clusterStats = jsonDecode(res.body));
      }
    } catch (_) {}
  }

  // ऑटो-डेबिट बैंक सेटलमेंट टेस्ट ट्रिगर
  Future<void> _triggerAutoDebitSettlement(String parentID) async {
    try {
      final url = '$kBackendBaseUrl/api/v1/payment/autopay-webhook?parent_id=$parentID&bank_utr=BANK_UTR_${DateTime.now().millisecondsSinceEpoch}&status=SUCCESS';
      final res = await http.get(Uri.parse(url));
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('ऑटो-डेबिट रिस्पॉन्स: ${res.body}')),
        );
      }
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0B0F19),
      appBar: AppBar(
        title: const Text('अनंत अभ्यास — 360° एडमिन क्लस्टर', style: TextStyle(color: Colors.white, fontSize: 18)),
        backgroundColor: const Color(0xFF111827),
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: Colors.amber,
          labelColor: Colors.amber,
          unselectedLabelColor: Colors.grey,
          tabs: const [
            Tab(icon: Icon(Icons.speed), text: 'टेलीमेट्री व ऑडिट'),
            Tab(icon: Icon(Icons.account_balance), text: 'पार्टनर लेजर Payout'),
            Tab(icon: Icon(Icons.cloud_sync), text: 'क्लस्टर व ऑटो-पे'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildTelemetryTab(),
          _buildPartnerLedgerTab(),
          _buildClusterHealthTab(),
        ],
      ),
    );
  }

  Widget _buildTelemetryTab() {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: _students.length,
      itemBuilder: (ctx, i) {
        final s = _students[i];
        return Card(
          color: const Color(0xFF1F2937),
          shape: RoundedRectangleBorder(
            side: BorderSide(color: s["flagged"] ? Colors.redAccent : Colors.transparent, width: 1.5),
            borderRadius: BorderRadius.circular(8),
          ),
          margin: const EdgeInsets.only(bottom: 12),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.between,
                  children: [
                    Text(s["name"], style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold)),
                    Chip(label: Text(s["wing"], style: const TextStyle(color: Colors.white, fontSize: 11)), backgroundColor: Colors.indigo),
                  ],
                ),
                Text('डेमो ID: ${s["demoID"]}', style: const TextStyle(color: Colors.grey)),
                Text('स्कोर: ${s["score"]} | गति: ${s["speed"]}', style: const TextStyle(color: Colors.amber, fontWeight: FontWeight.bold)),
                if (s["flagged"]) ...[
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.all(8),
                    color: Colors.red.withOpacity(0.2),
                    child: Text(
                      '⚠️ एंटी-चीट अलर्ट: दावा कक्षा ${s["claimed"]} का है लेकिन क्षमता कक्षा ${s["detected"]} पाई गई!',
                      style: const TextStyle(color: Colors.redAccent, fontSize: 12, fontWeight: FontWeight.bold),
                    ),
                  ),
                ],
                const SizedBox(height: 10),
                ElevatedButton.icon(
                  icon: const Icon(Icons.card_membership, size: 16),
                  label: const Text('सर्टिफिकेट लिंक देखें'),
                  style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF374151), foregroundColor: Colors.white),
                  onPressed: () {
                    showDialog(
                      context: context,
                      builder: (ctx) => AlertDialog(
                        title: const Text('डिजिटल सर्टिफिकेट URL'),
                        content: SelectableText('$kBackendBaseUrl/cert/${s["demoID"]}'),
                        actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('बंद करें'))],
                      ),
                    );
                  },
                )
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildPartnerLedgerTab() {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: _partnerLedger.length,
      itemBuilder: (ctx, i) {
        final p = _partnerLedger[i];
        final bool settled = p["isSettled"];
        return Card(
          color: const Color(0xFF1F2937),
          margin: const EdgeInsets.only(bottom: 12),
          child: ListTile(
            title: Text(p["shop"], style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
            subtitle: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('कोड: ${p["code"]} | कुल पेड छात्र: ${p["conversions"]}', style: const TextStyle(color: Colors.grey)),
                Text('देय कमीशन: ₹${p["payout"]} | UPI: ${p["upi"]}', style: const TextStyle(color: Colors.greenAccent)),
              ],
            ),
            trailing: settled
                ? const Chip(label: Text('Paid'), backgroundColor: Colors.teal)
                : ElevatedButton(
                    style: ElevatedButton.styleFrom(backgroundColor: Colors.amber, foregroundColor: Colors.black),
                    onPressed: () {
                      setState(() => p["isSettled"] = true);
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(content: Text('${p["shop"]} का ₹${p["payout"]} पेआउट सेटल हुआ!')),
                      );
                    },
                    child: const Text('UPI Pay'),
                  ),
          ),
        );
      },
    );
  }

  Widget _buildClusterHealthTab() {
    return Padding(
      padding: const EdgeInsets.all(20.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('क्लस्टर स्थिति (Backend Telemetry)', style: TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 15),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(color: const Color(0xFF1F2937), borderRadius: BorderRadius.circular(8)),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('सिस्टम: ${_clusterStats["system"] ?? "Anant Abhyas Ultra Core"}', style: const TextStyle(color: Colors.white)),
                const SizedBox(height: 6),
                Text('क्लस्टर स्टेट: ${_clusterStats["cluster_state"] ?? "ACTIVE"}', style: const TextStyle(color: Colors.greenAccent)),
                const SizedBox(height: 6),
                Text('ऑटो-पे नीति: ${_clusterStats["auto_pay"] ?? "ENFORCED_MANDATE_ONLY"}', style: const TextStyle(color: Colors.amber)),
              ],
            ),
          ),
          const SizedBox(height: 30),
          const Text('ऑटो-डेबिट सिमुलेशन (UPI AutoPay Test)', style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold)),
          const SizedBox(height: 10),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.indigo, foregroundColor: Colors.white),
            onPressed: () => _triggerAutoDebitSettlement("DEMO-2026-4973"),
            child: const Text('बैंक ऑटो-डेबिट वेबहुक ट्रिगर करें (Test Unlock)'),
          )
        ],
      ),
    );
  }
}

