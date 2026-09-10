import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const String kBackendBaseUrl = "http://YOUR_SERVER_IP:8080";

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
    _tabController = TabController(length: 2, vsync: this);
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
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildTelemetryTab(),
          _buildPartnerLedgerTab(),
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
                  
mainAxisAlignment: MainAxisAlignment.spaceBetween,

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
                Text('कोड: ${p["code"]} | कुल छात्र: ${p["conversions"]}', style: const TextStyle(color: Colors.grey)),
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
}

