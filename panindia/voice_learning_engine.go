package panindia

import (
	"strings"
)

// VoiceLearningEngine यह ट्रैक करता है कि यूजर कैसे बात कर रहा है और AI को उसके मुताबिक कैसे ढालना है
type VoiceLearningEngine struct {
	MasterEngine *MasterPanIndiaEngine
}

func NewVoiceLearningEngine() *VoiceLearningEngine {
	return &VoiceLearningEngine{
		MasterEngine: NewMasterPanIndiaEngine(),
	}
}

// AdaptiveResponseConfig यह तय करता है कि सामने वाले को जवाब देते समय AI की आवाज और भाषा का स्टाइल क्या होगा
type AdaptiveResponseConfig struct {
	TargetLanguage    string `json:"target_language"`
	SelectedAccent    string `json:"selected_accent"`
	SpeechPacing      string `json:"speech_pacing"`       // जैसे: Rapid, Moderate, Slow-Melodic
	HumanConversationalTone string `json:"human_tone"`  // जैसे: Warm teacher, Energetic brother
	IsAdaptiveMatch   bool   `json:"is_adaptive_match"`
}

// LearnAndAdaptToUser: 
// यह फंक्शन यूजर के इनपुट को सुनता/पढ़ता है, उसका लहजा और भाषा सीखता है, 
// और उसी के आधार पर AI के लिए रिस्पॉन्स कॉन्फिगरेशन तैयार करता है।
func (v *VoiceLearningEngine) LearnAndAdaptToUser(userRawInput string) AdaptiveResponseConfig {
	cleanedInput := strings.ToLower(strings.TrimSpace(userRawInput))

	// मास्टर इंजन से वॉयस और एक्सेंट प्रोफाइल मंगाओ
	profile := v.MasterEngine.DetectVoiceAndAccentProfile(cleanedInput)

	// यहाँ सिस्टम यूजर के लहजे को सीखकर उसके अनुसार रिस्पॉन्स पैरामीटर्स सेट करता है
	isMatch := false
	if cleanedInput != "" && profile.StateName != "Pan-India Fallback" {
		isMatch = true
	}

	return AdaptiveResponseConfig{
		TargetLanguage:          profile.LanguageName,
		SelectedAccent:          profile.Accent,
		SpeechPacing:            profile.SpeechCadence,
		HumanConversationalTone: profile.ToneStyle,
		IsAdaptiveMatch:         isMatch,
	}
}

// GenerateHumanLikeAIInstructions: 
// यह फंक्शन AI को सख्त निर्देश देता है कि उसे रोबोट की तरह नहीं, बल्कि एक असली इंसान की तरह 
// उसी क्षेत्रीय अंदाज और गति में बात करनी है।
func (v *VoiceLearningEngine) GenerateHumanLikeAIInstructions(config AdaptiveResponseConfig) string {
	return "=== मानव-समान संवाद और एक्सेंट निर्देश (Human-Like Voice Engine) ===\n" +
		"1. भाषा और एक्सेंट: " + config.TargetLanguage + " (" + config.SelectedAccent + ")\n" +
		"2. बोलने की गति (Pacing): " + config.SpeechPacing + "\n" +
		"3. मानवीय टोन (Tone): " + config.HumanConversationalTone + "\n" +
		"4. नियम: कभी भी यह अहसास न होने दें कि सामने AI बैठा है। ऐसे बात करें जैसे कोई अनुभवी और अपना शिक्षक या बड़ा भाई सामने बैठकर समझा रहा हो। क्षेत्रीय बोलचाल के प्राकृतिक ठहराव (Natural Pauses) का उपयोग करें।"
}
