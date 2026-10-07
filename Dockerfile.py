# पाइथन का हल्का और तेज ऑफिशियल बेस वर्जन उपयोग करें
FROM python:3.10-slim

# वर्किंग डायरेक्टरी सेट करें
WORKDIR /app

# सिस्टम की जरूरी लाइब्रेरीज और डिपेंडेंसीज इंस्टॉल करें (Faster-Whisper और SQLite के लिए)
RUN apt-get update && apt-get install -y \
    build-essential \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# requirements.txt को कॉपी करें और इंस्टॉल करें
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# प्रोजेक्ट की बाकी सारी फाइलें (voice_service.py, knowledge_core.py आदि) कॉपी करें
COPY . .

# पोर्ट्स ओपन करें (8000: Voice Service, 8001: Knowledge Core)
EXPOSE 8000
EXPOSE 8001

# डिफ़ॉल्ट कमांड: चूंकि हमारे पास दो अलग-अलग माइक्रोसर्विसेज हैं, 
# हम Render या Hetzner पर डिप्लॉयमेंट के वक्त स्टार्टअप कमांड सेट करेंगे।
# (जैसे: uvicorn voice_service:app --host 0.0.0.0 --port 8000)
CMD ["uvicorn", "voice_service:app", "--host", "0.0.0.0", "--port", "8000"]
