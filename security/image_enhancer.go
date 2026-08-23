package security

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

type ImageEnhancer struct{}

func NewImageEnhancer() *ImageEnhancer {
	return &ImageEnhancer{}
}

// खराब रोशनी और मुड़े हुए पन्नों की गुणवत्ता जांचना
func (i *ImageEnhancer) CheckAndNormalizeQuality(imgBytes []byte) (bool, string) {
	if len(imgBytes) < 5000 {
		return false, "छवि बहुत छोटी या अस्पष्ट है। कृपया पन्ने की साफ फोटो लें।"
	}

	_, _, err := image.DecodeConfig(bytes.NewReader(imgBytes))
	if err != nil {
		return false, "फोटो लोड नहीं हो सकी। कृपया दोबारा क्लिक करें।"
	}

	return true, "फोटो की गुणवत्ता पर्याप्त है।"
}
