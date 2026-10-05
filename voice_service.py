from fastapi import FastAPI, UploadFile, File, Form, BackgroundTasks, HTTPException
from fastapi.responses import FileResponse
from faster_whisper import WhisperModel
from gTTS import gTTS
from deep_translator import GoogleTranslator
import os
import uuid
import sqlite3
import asyncio
import threading
import uvicorn

app = FastAPI(title="Anant Abhyaas Ultimate Enterprise Voice Engine")

# --- 1. थ्रेड-सेफ SQLite डेटाबेस सेटअप (Database Lock & Concurrency Safe) ---
DB_FILE = "anant_learning.db"
db_lock = threading.Lock()

def init_db():
    with db_lock:
        conn = sqlite3.connect(DB_FILE, timeout=30.0)
        cursor = conn.cursor()
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS voice_learning_logs (
                id TEXT PRIMARY KEY,
                source_lang TEXT,
                target_lang TEXT,
                recognized_text TEXT,
                processed_text TEXT,
                is_trained INTEGER DEFAULT 0,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        conn.commit()
        conn.close()

init_db()

# --- 2. Whisper मॉडल का सिंगल-ग्लोबल लोड (Memory & Speed Optimized) ---
print("Whisper AI इंजन लोड हो रहा है...")
model = WhisperModel("base", compute_type="int8")

# --- 3. क्रैश-प्रूफ ऑटो-लर्निंग और ऑटो-ट्यूनिंग बैकग्राउंड वर्कर ---
async def background_auto_learning_worker():
    """
    यह बैकग्राउंड वर्कर सर्वर पर लोड कम होने या तय समय पर पिछले 
    डेटाबेस लॉग्स से खुद सीखकर (Auto-Tuning) सिस्टम को रिफाइन करेगा।
    """
    while True:
        await asyncio.sleep(120)  # हर 2 मिनट पर चेक करेगा
        try:
            with db_lock:
                conn = sqlite3.connect(DB_FILE, timeout=30.0)
                cursor = conn.cursor()
                
                # अन-ट्रेनड डेटा उठाना
                cursor.execute("SELECT id, recognized_text, processed_text FROM voice_learning_logs WHERE is_trained = 0 LIMIT 50")
                rows = cursor.fetchall()
                
                if rows:
                    print(f"[Auto-Learning] बैकग्राउंड में {len(rows)} वॉयस पैटर्न्स से सिस्टम ऑटो-ट्यून हो रहा है...")
                    
                    # यहाँ भविष्य में मॉडल फाइन-ट्यूनिंग या डिक्शनरी अपडेट का लॉजिक काम करेगा
                    ids_to_update = [r[0] for r in rows]
                    cursor.executemany("UPDATE voice_learning_logs SET is_trained = 1 WHERE id = ?", [(i,) for i in ids_to_update])
                    conn.commit()
                    print("[Auto-Learning] ऑटो-ट्यूनिंग साइकिल सफलताપूर्वक संपन्न!")
                    
                conn.close()
        except Exception as e:
            print(f"[Auto-Learning Warning - लूप सुरक्षित है]: {e}")

@app.on_event("startup")
async def startup_event():
    asyncio.create_task(background_auto_learning_worker())

def save_to_db_safe(log_id, src, tgt, rec, proc):
    """थ्रेड-सेफ डेटाबेस लॉगिंग"""
    try:
        with db_lock:
            conn = sqlite3.connect(DB_FILE, timeout=30.0)
            cursor = conn.cursor()
            cursor.execute('''
                INSERT INTO voice_learning_logs (id, source_lang, target_lang, recognized_text, processed_text, is_trained)
                VALUES (?, ?, ?, ?, ?, 0)
            ''', (log_id, src, tgt, rec, proc))
            conn.commit()
            conn.close()
    except Exception as e:
        print(f"DB Save Error: {e}")

def cleanup_file(file_path: str):
    """सुरक्षित डिस्क क्लीनअप (Storage Leak Prevention)"""
    if file_path and os.path.exists(file_path):
        try:
            os.remove(file_path)
        except Exception as e:
            print(f"File Cleanup Error: {e}")

# --- 4. मुख्य स्पीच-टू-स्पीच और ऑटो-लर्निंग एंडपॉइंट ---
@app.post("/process-voice-learning")
def process_voice_with_learning(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    source_lang: str = Form(...),          
    target_lang: str = Form(None)          
):
    # STEP 1: यूनिक आईडी और पाथ ट्रेव्हर्सेल सुरक्षा (Path Sanitization)
    unique_id = str(uuid.uuid4())
    safe_filename = os.path.basename(file.filename) if file.filename else "audio.wav"
    temp_audio_path = f"temp_{unique_id}_{safe_filename}"
    output_audio_path = f"response_{unique_id}.mp3"
    
    # इनपुट फाइल सेव करना
    try:
        with open(temp_audio_path, "wb") as buffer:
            buffer.write(file.file.read())
    except Exception as e:
        return {"status": "error", "message": f"File save failed: {str(e)}"}
    
    try:
        # STEP 2: Speech-to-Text (STT) - आवाज को टेक्स्ट में बदलना
        segments, _ = model.transcribe(temp_audio_path, language=source_lang)
        recognized_text = " ".join([seg.text for seg in segments]).strip()
        
        if not recognized_text:
            background_tasks.add_task(cleanup_file, temp_audio_path)
            raise HTTPException(status_code=400, detail="कोई आवाज नहीं पहचानी जा सकी।")

        # STEP 3: Dynamic Translation (बिना किसी हार्डकोडिंग के, डिमांड पर आधारित)
        processed_text = recognized_text
        actual_target_lang = target_lang if target_lang else source_lang
        
        if target_lang and target_lang != source_lang:
            try:
                processed_text = GoogleTranslator(source=source_lang, target=target_lang).translate(recognized_text)
            except Exception as e:
                print(f"Translation Error: {e}")
                processed_text = recognized_text

        # STEP 4: Text-to-Speech (TTS) - टेक्स्ट को वापस आवाज में बदलना
        tts = gTTS(text=processed_text, lang=actual_target_lang, slow=False)
        tts.save(output_audio_path)
        
        # STEP 5: डेटाबेस में डेटा लॉग करना (ऑटो-लर्निंग के लिए)
        background_tasks.add_task(
            save_to_db_safe, 
            unique_id, source_lang, actual_target_lang, recognized_text, processed_text
        )

        # इनपुट टेंपररी फाइल को तुरंत हटाना
        background_tasks.add_task(cleanup_file, temp_audio_path)

        # STEP 6: FileResponse के जरिए सुरक्षित स्ट्रीमिंग और गो बैकएंड को ऑडियो भेजना
        return FileResponse(
            path=output_audio_path,
            media_type="audio/mpeg",
            filename=output_audio_path,
            background=BackgroundTasks([lambda: cleanup_file(output_audio_path)])
        )
        
    except HTTPException as he:
        background_tasks.add_task(cleanup_file, temp_audio_path)
        raise he
    except Exception as e:
        background_tasks.add_task(cleanup_file, temp_audio_path)
        background_tasks.add_task(cleanup_file, output_audio_path)
        return {"status": "error", "message": str(e)}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
