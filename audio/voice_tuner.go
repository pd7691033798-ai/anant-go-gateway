package audio

import (
	"fmt"
	"strings"
)

type VoiceTunerService struct {
	SpeakingRate string
	Pitch        string
}

func NewVoiceTunerService() *VoiceTunerService {
	return &VoiceTunerService{SpeakingRate: "88%", Pitch: "-1st"}
}

func (v *VoiceTunerService) SanitizePhonetics(rawText string) string {
	cleaned := rawText
	replacements := map[string]string{
		"घणो":      "घ-णो",
		"आछो":      "आ-छो",
		"टाबर":      "टा-बर",
		"फूटरी":     "फू-टरी",
		"ਬਹੁਤ ਵਧੀਆ": "ਬਹੁਤ ਵਧੀ-ਆ",
		"शाबाश":     "शाबाश!",
		"1/2":      "आधा",
		"3/4":      "तीन चौथाई",
	}
	for word, phonetic := range replacements {
		cleaned = strings.ReplaceAll(cleaned, word, phonetic)
	}
	return cleaned
}

func (v *VoiceTunerService) GenerateKidFriendlySSML(studentName, rawText, dialectTone string) string {
	phoneticText := v.SanitizePhonetics(rawText)
	return fmt.Sprintf(
		`<speak>
  <prosody rate="%s" pitch="%s">
    <emphasis level="moderate">%s, %s!</emphasis>
    <break time="450ms"/>
    %s
    <break time="300ms"/>
    <prosody rate="85%%">
      कल फिर 15 मिनट अभ्यास करेंगे। अपना ध्यान रखना!
    </prosody>
  </prosody>
</speak>`,
		v.SpeakingRate, v.Pitch, dialectTone, studentName, phoneticText,
	)
}
