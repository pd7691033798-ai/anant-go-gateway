from fastapi import FastAPI, UploadFile, File, Form, BackgroundTasks, HTTPException
from fastapi.responses import FileResponse
from faster_whisper import WhisperModel
from gTTS import gTTS
from deep_translator import GoogleTranslator
import os
import uuid
import sqlite3
import threading
import asyncio
import httpx
import uvicorn

app = FastAPI(title="Anant Abhyaas Redis-Ready Cloud Voice Gateway")

# --- 1. एब्स्ट्रैक्टेड डेटाबेस सेटअप ---
DB_URL = os.getenv("DATABASE_URL", "anant_voice_logs.db")
db_lock = threading.Lock()

VOICE_ARCHIVE_DIR = os.getenv("VOICE_ARCHIVE_DIR", "archived_voice_samples")
os.makedirs(VOICE_ARCHIVE_DIR, exist_ok=True)

def init_voice_db():
    with db_lock:
        conn = sqlite3.connect(DB_URL, timeout=30.0)
        cursor = conn.cursor()
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS voice_learning_logs (
                id TEXT PRIMARY KEY,
                source_lang TEXT,
                target_lang TEXT,
                recognized_text TEXT,
                processed_text TEXT,
                sample_path TEXT,
                is_trained INTEGER DEFAULT 0,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        conn.commit()
        conn.close()

init_voice_db()

# --- 2. Whisper मॉडल का क्लाउड लोड ---
print("Whisper AI क्लाउड इंजन लोड हो रहा है...")
model = WhisperModel("base", compute_type="int8")

KNOWLEDGE_CORE_URL = os.getenv("KNOWLEDGE_CORE_URL", "http://localhost:8001/get-knowledge")
http_client = httpx.AsyncClient(timeout=10.0)

@app.on_event("shutdown")
async def shutdown_event():
    await http_client.aclose()

# --- 3. REDIS-READY ABSTRACTED QUEUE INTERFACE ---
# यह लेयर आज Memory Queue की तरह काम कर रही है। भविष्य में इसे सीधे Redis backend से बदला जा सकता है।
class AbstractTaskQueue:
    def __init__(self, maxsize: int):
        self.queue = asyncio.Queue(maxsize=maxsize)
    
    async def put(self, item: dict, block: bool = True):
        if block:
            await self.queue.put(item)
        else:
            self.queue.put_nowait(item)
            
    async def get(self) -> dict:
        return await self.queue.get()
        
    def task_done(self):
        self.queue.task_done()
        
    def qsize(self) -> int:
        return self.queue.qsize()

MAX_QUEUE_SIZE = int(os.getenv("MAX_QUEUE_SIZE", 5000))
task_queue = AbstractTaskQueue(maxsize=MAX_QUEUE_SIZE)

# --- 4. सिंगल टास्क प्रोसेसिंग यूनिट ---
async def process_single_task(task_data):
    temp_audio_path = task_data["temp_audio_path"]
    output_audio_path = task_data["output_audio_path"]
    source_lang = task_data["source_lang"]
    desired_target_lang = task_data["desired_target_lang"]
    grade = task_data["grade"]
    stream = task_data["stream"]
    board = task_data["board"]
    external_api_url = task_data["external_api_url"]
    unique_id = task_data["unique_id"]
    future = task_data["future"]

    archived_sample_path = None

    try:
        # STEP 1: वॉयस सैंपल आर्काइव करना (फ्यूचर ट्रेनिंग के लिए डेटा बैंक)
        safe_ext = os.path.splitext(temp_audio_path)[1] or ".wav"
        archived_sample_path = os.path.join(VOICE_ARCHIVE_DIR, f"sample_{unique_id}{safe_ext}")
        if os.path.exists(temp_audio_path):
            with open(temp_audio_path, "rb") as src_file:
                with open(archived_sample_path, "wb") as dest_file:
                    dest_file.write(src_file.read())

        # STEP 2: Speech-to-Text (Faster-Whisper - तमिल, तेलुगु या कोई भी क्षेत्रीय भाषा)
        segments, info = await asyncio.to_thread(
            model.transcribe, temp_audio_path, language=None if source_lang == "auto" else source_lang
        )
        recognized_text = " ".join([seg.text for seg in segments]).strip()
        detected_source_lang = info.language if source_lang == "auto" else source_lang
        
        if not recognized_text:
            future.set_exception(HTTPException(status_code=400, detail="कोई आवाज नहीं पहचानी जा सकी।"))
            return

        # STEP 3: Knowledge Core से संपर्क (Cloud Microservice Bridge)
        knowledge_answer = f"[{recognized_text}] से संबंधित अकादमिक जानकारी।"
        try:
            res = await http_client.post(KNOWLEDGE_CORE_URL, json={
                "grade": grade, "stream": stream, "board": board, 
                "query": recognized_text, "external_api_url": external_api_url
            })
            if res.status_code == 200:
                knowledge_answer = res.json().get("content", knowledge_answer)
        except Exception as ex:
            print(f"Cloud Knowledge Bridge Warning: {ex}")

        # STEP 4: Dynamic Polyglot Translation (छात्र की मांगी गई भाषा में बदलना)
        target_lang_code = desired_target_lang if desired_target_lang else detected_source_lang
        final_response_text = knowledge_answer
        
        if target_lang_code and target_lang_code != detected_source_lang:
            try:
                final_response_text = await asyncio.to_thread(
                    GoogleTranslator(source='auto', target=target_lang_code).translate, knowledge_answer
                )
            except Exception as e:
                print(f"Dynamic Translation Error: {e}")

        # STEP 5: Text-to-Speech (आवाज बनाना)
        await asyncio.to_thread(
            lambda: gTTS(text=final_response_text, lang=target_lang_code, slow=False).save(output_audio_path)
        )
        
        # STEP 6: डेटाबेस लॉगिंग
        try:
            with db_lock:
                conn = sqlite3.connect(DB_URL, timeout=30.0)
                cursor = conn.cursor()
                cursor.execute('''
                    INSERT INTO voice_learning_logs (id, source_lang, target_lang, recognized_text, processed_text, sample_path, is_trained)
                    VALUES (?, ?, ?, ?, ?, ?, 0)
                ''', (unique_id, detected_source_lang, target_lang_code, recognized_text, final_response_text, archived_sample_path))
                conn.commit()
                conn.close()
        except Exception as db_ex:
            print(f"Cloud DB Save Error: {db_ex}")

        future.set_result(output_audio_path)

    except Exception as e:
        if not future.done():
            future.set_exception(e)
    finally:
        if os.path.exists(temp_audio_path):
            try:
                os.remove(temp_audio_path)
            except:
                pass

# --- 5. ऑटोनॉमस एडाप्टिव लोड डिस्पैचर ---
async def adaptive_load_dispatcher():
    while True:
        try:
            task = await task_queue.get()
            asyncio.create_task(process_single_task(task))
            task_queue.task_done()
        except Exception as e:
            print(f"[Dispatcher Error]: {e}")

@app.on_event("startup")
async def startup_event():
    asyncio.create_task(adaptive_load_dispatcher())

# --- 6. मुख्य पॉलीग्लॉट वॉयस एंडपॉइंट (Go Backend से कनेक्ट होने के लिए) ---
@app.post("/polyglot-voice-query")
async def polyglot_voice_query(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    grade: str = Form("Class_12"),
    stream: str = Form("General"),
    board: str = Form("CBSE"),
    source_lang: str = Form("auto"),
    desired_target_lang: str = Form("hi"),
    external_api_url: str = Form(None)
):
    unique_id = str(uuid.uuid4())
    safe_filename = os.path.basename(file.filename) if file.filename else "audio.wav"
    temp_audio_path = f"temp_{unique_id}_{safe_filename}"
    output_audio_path = f"response_{unique_id}.mp3"
    
    try:
        with open(temp_audio_path, "wb") as buffer:
            buffer.write(await file.read())
    except Exception as e:
        return {"status": "error", "message": f"File save failed: {str(e)}"}
    
    future = asyncio.get_running_loop().create_future()

    task_payload = {
        "unique_id": unique_id,
        "temp_audio_path": temp_audio_path,
        "output_audio_path": output_audio_path,
        "grade": grade,
        "stream": stream,
        "board": board,
        "source_lang": source_lang,
        "desired_target_lang": desired_target_lang,
        "external_api_url": external_api_url,
        "future": future
    }

    try:
        await task_queue.put(task_payload, block=False)
    except asyncio.QueueFull:
        if os.path.exists(temp_audio_path):
            os.remove(temp_audio_path)
        raise HTTPException(status_code=503, detail="क्लाउड सर्वर पर अत्यधिक ट्रैफिक है, कृपया कुछ सेकंड बाद प्रयास करें।")

    try:
        processed_audio = await future
        return FileResponse(
            path=processed_audio,
            media_type="audio/mpeg",
            filename=processed_audio,
            background=BackgroundTasks([lambda: os.path.exists(processed_audio) and os.remove(processed_audio)])
        )
    except HTTPException as he:
        raise he
    except Exception as e:
        if os.path.exists(output_audio_path):
            try:
                os.remove(output_audio_path)
            except:
                pass
        return {"status": "error", "message": str(e)}

if __name__ == "__main__":
    port = int(os.getenv("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
                         
