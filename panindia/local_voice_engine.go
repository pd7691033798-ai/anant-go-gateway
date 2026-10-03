package panindia

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VoiceSampleData हर राज्य की भाषा, उसके सैंपल फाइल पाथ और बोलने के तरीके (Cadence/Tone) को स्टोर करता है
type VoiceSampleData struct {
	StateName    string   `json:"state_name"`
	LanguageName string   `json:"language_name"`
	Accent       string   `json:"accent"`
	SpeechCadence string  `json:"speech_cadence"`
	SampleAudioPath string `json:"sample_audio_path"` // यहाँ तू रोज अपनी सैंपल ऑडियो फाइल का पाथ देगा
	SampleKeywords []string `json:"sample_keywords"`  // इनेंट (Intent) मैच करने के लिए कीवर्ड्स
}

type LocalVoiceEngine struct {
	BaseAssetsDir string
	RegistryFile  string
}

func NewLocalVoiceEngine(assetsDir string) *LocalVoiceEngine {
	return &LocalVoiceEngine{
		BaseAssetsDir: assetsDir,
		RegistryFile:  filepath.Join(assetsDir, "voice_registry.json"),
	}
}

// 1. रोज नया सैंपल जोड़ने या अपडेट करने का फंक्शन (Daily Sample Adder)
func (lve *LocalVoiceEngine) AddOrUpdateDailySample(data VoiceSampleData) error {
	// डायरेक्टरी बनाएं यदि न हो
	if err := os.MkdirAll(lve.BaseAssetsDir, 0755); err != nil {
		return err
	}

	// मौजूदा डेटा लोड करें
	registry := make(map[string]VoiceSampleData)
	if fileData, err := os.ReadFile(lve.RegistryFile); err == nil {
		_ = json.Unmarshal(fileData, &registry)
	}

	// नया या अपडेटेड सैंपल जोड़ें (State/Language के हिसाब से)
	key := strings.ToLower(data.StateName + "_" + data.LanguageName)
	registry[key] = data

	// वापस JSON फाइल में सेव करें
	updatedBytes, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(lve.RegistryFile, updatedBytes, 0644)
}

// 2. स्पीच-टू-टेक्स्ट और इंटेंट मैचिंग (STT Simulation & Intent Recognition)
// यह चेक करता है कि बच्चा क्या पूछना चाहता है और कौन सी भाषा/राज्य का है
func (lve *LocalVoiceEngine) MatchIntentAndLanguage(userInputOrAudioText string) (VoiceSampleData, string, bool) {
	cleanedInput := strings.ToLower(strings.TrimSpace(userInputOrAudioText))

	// रजिस्ट्री फाइल से सारे सैंपल्स पढ़ें
	fileData, err := os.ReadFile(lve.RegistryFile)
	if err != nil {
		return VoiceSampleData{}, "डिफ़ॉल्ट हिंदी फॉलबैक", false
	}

	registry := make(map[string]VoiceSampleData)
	_ = json.Unmarshal(fileData, &registry)

	// बच्चे के इनपुट में से कीवर्ड्स और राज्य की भाषा मैच करना
	for _, sample := range registry {
		// राज्य या भाषा का नाम मैच हो रहा है या नहीं
		if strings.Contains(cleanedInput, strings.ToLower(sample.StateName)) || 
		   strings.Contains(cleanedInput, strings.ToLower(sample.LanguageName)) {
			
			// अब इंटेंट (काम/चैप्टर पढ़ाने की बात) चेक करें
			for _, kw := range sample.SampleKeywords {
				if strings.Contains(cleanedInput, strings.ToLower(kw)) {
					responseMsg := fmt.Sprintf("राज्य: %s | भाषा: %s | एक्सेंट: %s के आधार पर सैंपल मिल गया। अब इसी लहजे में चैप्टर पढ़ाने की कार्रवाई शुरू की जा रही है।", sample.StateName, sample.LanguageName, sample.Accent)
					return sample, responseMsg, true
				}
			}
			return sample, fmt.Sprintf("भाषा मैच हो गई: %s, लेकिन स्पेसिफिक कीवर्ड नहीं मिला। डिफ़ॉल्ट सैंपल से काम चलाया जा रहा है।", sample.LanguageName), true
		}
	}

	// यदि कुछ मैच न हो तो डिफ़ॉल्ट हिंदी सैंपल उठाओ
	if defaultSample, exists := registry["uttar pradesh_hindi"]; exists {
		return defaultSample, "कोई क्षेत्रीय भाषा मैच नहीं हुई। डिफ़ॉल्ट हिंदी सैंपल से उत्तर दिया जा रहा है.", false
	}

	return VoiceSampleData{StateName: "Default", LanguageName: "Hindi", Accent: "Standard"}, "फॉलबैक मोड", false
}

// 3. टेक्स्ट-टू-स्पीच (TTS Simulation via Local Audio Samples)
// यह लोकल सैंपल से टोन और आवाज़ का स्टाइल उठाकर आउटपुट तैयार करता है
func (lve *LocalVoiceEngine) SynthesizeVoiceFromSample(sample VoiceSampleData, textToDeliver string) (string, string) {
	audioPath := sample.SampleAudioPath
	if audioPath == "" {
		audioPath = filepath.Join(lve.BaseAssetsDir, "default_audio.wav")
	}

	// यहाँ सिस्टम सैंपल फाइल को देखकर उसकी टोन और पेसिंग लागू कर देता है
	executionLog := fmt.Sprintf("[TTS Local Engine] सैंपल फाइल का उपयोग किया गया: %s | पेसिंग/लहजा: %s | डिलीवर होने वाला टेक्स्ट: %s", audioPath, sample.SpeechCadence, textToDeliver)
	
	return audioPath, executionLog
}
