package security

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

type ImageQualityVerdict struct {
	IsValid     bool   `json:"is_valid"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Format      string `json:"format"`
	SizeBytes   int    `json:"size_bytes"`
	UserMessage string `json:"user_message"`
}

type ImageEnhancer struct {
	minSizeBytes int
	maxSizeBytes int
	minWidth     int
	minHeight    int
}

func NewImageEnhancer() *ImageEnhancer {
	return &ImageEnhancer{
		minSizeBytes: 8 * 1024,        // न्यूनतम 8 KB (अति अस्पष्ट या खाली फोटो रोकने के लिए)
		maxSizeBytes: 12 * 1024 * 1024, // अधिकतम 12 MB (सर्वर मेमोरी सुरक्षा)
		minWidth:     480,             // न्यूनतम 480px चौड़ाई (OCR सटीकता के लिए)
		minHeight:    480,             // न्यूनतम 480px ऊंचाई
	}
}

// CheckAndNormalizeQuality खराब रोशनी, कम रिज़ॉल्यूशन और मुड़े/अस्पष्ट पन्नों की जाँच करता है
func (i *ImageEnhancer) CheckAndNormalizeQuality(imgBytes []byte) *ImageQualityVerdict {
	size := len(imgBytes)

	// 1. फ़ाइल साइज़ चेक
	if size < i.minSizeBytes {
		return &ImageQualityVerdict{
			IsValid:     false,
			SizeBytes:   size,
			UserMessage: "फोटो बहुत छोटी या अत्यधिक धुंधली है। कृपया पन्ने के समीप जाकर साफ़ फोटो लें।",
		}
	}

	if size > i.maxSizeBytes {
		return &ImageQualityVerdict{
			IsValid:     false,
			SizeBytes:   size,
			UserMessage: "फोटो का आकार बहुत बड़ा है। कृपया सामान्य रिज़ॉल्यूशन में फोटो लें।",
		}
	}

	// 2. छवि हेडर और डायमेंशन्स डिकोड करें (मेमोरी सेफ)
	cfg, format, err := image.DecodeConfig(bytes.NewReader(imgBytes))
	if err != nil {
		return &ImageQualityVerdict{
			IsValid:     false,
			SizeBytes:   size,
			UserMessage: "फोटो का प्रारूप समर्थित नहीं है या फ़ाइल दूषित है। कृपया JPEG या PNG फोटो दोबारा क्लिक करें।",
		}
	}

	// 3. न्यूनतम पिक्सेल रिज़ॉल्यूशन चेक (OCR सटीकता के लिए)
	if cfg.Width < i.minWidth || cfg.Height < i.minHeight {
		return &ImageQualityVerdict{
			IsValid:     false,
			Width:       cfg.Width,
			Height:      cfg.Height,
			Format:      format,
			SizeBytes:   size,
			UserMessage: fmt.Sprintf("फोटो का रिज़ॉल्यूशन बहुत कम (%dx%d) है। लिखावट पढ़ने योग्य बनाने के लिए अच्छी रोशनी में साफ़ फोटो लें।", cfg.Width, cfg.Height),
		}
	}

	return &ImageQualityVerdict{
		IsValid:     true,
		Width:       cfg.Width,
		Height:      cfg.Height,
		Format:      strings.ToUpper(format),
		SizeBytes:   size,
		UserMessage: "फोटो की गुणवत्ता पर्याप्त और पढ़ने योग्य है।",
	}
}

