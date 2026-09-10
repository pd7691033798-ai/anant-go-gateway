import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const String kBackendBaseUrl = "http://localhost:8080"; // अपना सर्वर URL डालें

void main() {
  FlutterError.onError = (FlutterErrorDetails details) {
    // बैकएंड ऑटो-हीलिंग क्रैश हुक पर एरर भेजना
    http.post(
      Uri.parse('$kBackendBaseUrl/api/v1/app/support-hook?parent_id=APP_CRASH&error_log=${Uri.encodeComponent(details.exceptionAsString())}'),
    );
  };
  runApp(const MaterialApp(
    debugShowCheckedModeBanner: false,
    home: StudentKioskApp(),
  ));
}

class StudentKioskApp extends StatefulWidget {
  const StudentKioskApp({super.key});

  @override
  State<StudentKioskApp> createState() => _StudentKioskAppState();
}

class _StudentKioskAppState extends State<StudentKioskApp> {
  final String studentID = "DEMO-2026-4973";
  final String masterPin = "982143";

  int _sessionSecondsLeft = 900; // 15 मिनट
  int _lockSecondsLeft = 180;    // 3 मिनट
  bool _isReportLocked = false;
  int _cooldownSeconds = 0;
  int _currentQIdx = 0;
  int _score = 0;

  Timer? _sessionTimer;
  Timer? _pingTimer;
  Timer? _lockTimer;
  Timer? _cooldownTimer;

  final List<Map<String, dynamic>> _cbtQuestions = [
    {
      "qId": "q1",
      "question": "यदि f(x) = x² + 2x + 1 है, तो f'(2) का मान क्या होगा? (JEE CBT)",
      "options": ["4", "5", "6", "8"],
      "correct": "C"
    },
    {
      "qId": "q2",
      "question": "एक रेलगाड़ी 60 किमी/घंटा से चलती है। 3 घंटे में तय दूरी? (नवोदय विंग)",
      "options": ["120 किमी", "180 किमी", "240 किमी", "150 किमी"],
      "correct": "B"
    },
    {
      "qId": "q3",
      "question": "प्रकाश संश्लेषण हेतु पौधों को कौन सी गैस आवश्यक है?",
      "options": ["ऑक्सीजन", "नाइट्रोजन", "कार्बन डाइऑक्साइड", "हाइड्रोजन"],
      "correct": "C"
    }
  ];

  @override
  void initState() {
    super.initState();
    _startSession();
    _startAggregatorPings();
  }

  // 1. 30-सेकंड पिंग बफर इंजन (Aggregator Engine)
  void _startAggregatorPings() {
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (timer) async {
      try {
        await http.get(Uri.parse('$kBackendBaseUrl/api/v1/ping?student_id=$studentID'));
      } catch (_) {}
    });
  }

  // 2. 15-मिनट Kiosk टाइमर
  void _startSession() {
    _sessionTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_sessionSecondsLeft > 0) {
        setState(() => _sessionSecondsLeft--);
      } else {
        _sessionTimer?.cancel();
        _lockSessionReport();
      }
    });
  }

  // 3. 3-सेकंड नो-तुक्का एंटी-चीट इंजन
  void _triggerAntiCheatCooldown() {
    setState(() => _cooldownSeconds = 3);
    _cooldownTimer?.cancel();
    _cooldownTimer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (_cooldownSeconds > 0) {
        setState(() => _cooldownSeconds--);
      } else {
        _cooldownTimer?.cancel();
      }
    });
  }

  void _onAnswerSelected(int index) {
    if (_cooldownSeconds > 0) return;

    final optChar = String.fromCharCode(65 + index); // A, B, C, D
    if (optChar == _cbtQuestions[_currentQIdx]["correct"]) {
      _score += (20 ~/ _cbtQuestions.length);
    }

    _triggerAntiCheatCooldown();

    Future.delayed(const Duration(milliseconds: 500), () {
      if (_currentQIdx < _cbtQuestions.length - 1) {
        setState(() => _currentQIdx++);
      } else {
        _submitCBTTest();
      }
    });
  }

  // 4. CBT सबमिशन बैकएंड को भेजना
  Future<void> _submitCBTTest() async {
    _sessionTimer?.cancel();
    try {
      final payload = {
        "track_code": "IIT_JEE",
        "student_id": studentID,
        "answers": {"q1": "C", "q2": "B", "q3": "C"}
      };
      await http.post(
        Uri.parse('$kBackendBaseUrl/api/v1/cbt/submit'),
        headers: {"Content-Type": "application/json"},
        body: jsonEncode(payload),
      );
    } catch (_) {}
    _lockSessionReport();
  }

  // 5. 3-मिनट रिपोर्ट लॉक
  void _lockSessionReport() {
    setState(() => _isReportLocked = true);
    _lockTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_lockSecondsLeft > 0) {
        setState(() => _lockSecondsLeft--);
      } else {
        _lockTimer?.cancel();
      }
    });
  }

  // 6. 100% UPI ऑटो-पे मैंडेट इनीशियलाइज़र
  Future<void> _initiateAutoPay() async {
    try {
      final res = await http.get(Uri.parse('$kBackendBaseUrl/api/v1/payment/setup-autopay?parent_id=$studentID&tier=PRO_699&exam_addon=true'));
      final data = jsonDecode(res.body);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('ऑटो-पे मैंडेट शुरू: ₹${data["monthly_amount"]}/माह (${data["message"]})')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('ऑटो-पे संपर्क विफल, ऑफ़लाइन मोड सक्रिय।')),
        );
      }
    }
  }

  void _showMasterPinDialog() {
    final pinController = TextEditingController();
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => AlertDialog(
        title: const Text('अभिभावक मास्टर PIN'),
        content: TextField(
          controller: pinController,
          keyboardType: TextInputType.number,
          maxLength: 6,
          obscureText: true,
          decoration: const InputDecoration(hintText: '6-अंकीय PIN लिखें'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('रद्द')),
          ElevatedButton(
            onPressed: () {
              if (pinController.text == masterPin) {
                Navigator.pop(ctx);
                setState(() {
                  _isReportLocked = false;
                  _sessionSecondsLeft = 900;
                });
                _startSession();
              } else {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('गलत PIN!')),
                );
              }
            },
            child: const Text('अनलॉक करें'),
          )
        ],
      ),
    );
  }

  @override
  void dispose() {
    _sessionTimer?.cancel();
    _pingTimer?.cancel();
    _lockTimer?.cancel();
    _cooldownTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: Scaffold(
        backgroundColor: const Color(0xFF0F172A),
        appBar: AppBar(
          title: const Text('अनंत अभ्यास (Student Kiosk CBT)', style: TextStyle(color: Colors.white, fontSize: 16)),
          backgroundColor: const Color(0xFF1E293B),
          automaticallyImplyLeading: false,
          actions: [
            Center(
              child: Padding(
                padding: const EdgeInsets.only(right: 16.0),
                child: Text(
                  '⏱️ ${_sessionSecondsLeft ~/ 60}:${(_sessionSecondsLeft % 60).toString().padLeft(2, '0')}',
                  style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.amber),
                ),
              ),
            )
          ],
        ),
        body: _isReportLocked ? _buildReportLockUI() : _buildCBTInterface(),
      ),
    );
  }

  Widget _buildCBTInterface() {
    final q = _cbtQuestions[_currentQIdx];
    return Padding(
      padding: const EdgeInsets.all(20.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          LinearProgressIndicator(
            value: (_currentQIdx + 1) / _cbtQuestions.length,
            color: Colors.amber,
            backgroundColor: Colors.grey.shade800,
          ),
          const SizedBox(height: 20),
          Text('प्रश्न ${_currentQIdx + 1} / ${_cbtQuestions.length}', style: const TextStyle(color: Colors.grey)),
          const SizedBox(height: 10),
          Text(q["question"], style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 20),
          if (_cooldownSeconds > 0)
            Container(
              padding: const EdgeInsets.all(8),
              color: Colors.red.shade900,
              child: Text('🚫 3-सेकंड नो-तुक्का लॉक: ${_cooldownSeconds}s', textAlign: TextAlign.center, style: const TextStyle(color: Colors.white)),
            ),
          const SizedBox(height: 10),
          ...List.generate(q["options"].length, (idx) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 12.0),
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF1E293B),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                ),
                onPressed: _cooldownSeconds > 0 ? null : () => _onAnswerSelected(idx),
                child: Text("${String.fromCharCode(65 + idx)}. ${q["options"][idx]}", style: const TextStyle(fontSize: 16)),
              ),
            );
          }),
        ],
      ),
    );
  }

  Widget _buildReportLockUI() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.verified, color: Colors.greenAccent, size: 70),
            const SizedBox(height: 16),
            const Text('15-मिनट अभ्यास सत्र पूर्ण!', style: TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            Text('कुल प्राप्तांक: $_score / 20 अंक', style: const TextStyle(color: Colors.amber, fontSize: 18)),
            const SizedBox(height: 16),
            Text('रिपोर्ट लॉक टाइमर: ${_lockSecondsLeft ~/ 60}:${(_lockSecondsLeft % 60).toString().padLeft(2, '0')}', style: const TextStyle(color: Colors.redAccent, fontSize: 18)),
            const SizedBox(height: 30),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: Colors.amber, foregroundColor: Colors.black),
              onPressed: _showMasterPinDialog,
              child: const Text('अभिभावक मास्टर PIN दर्ज करें'),
            ),
            const SizedBox(height: 15),
            OutlinedButton(
              style: OutlinedButton.styleFrom(foregroundColor: Colors.tealAccent, side: const BorderSide(color: Colors.tealAccent)),
              onPressed: _initiateAutoPay,
              child: const Text('UPI ऑटो-पे (₹1,098/माह) एक्टिवेट करें'),
            )
          ],
        ),
      ),
    );
  }
}
