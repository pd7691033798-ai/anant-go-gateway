package stealth

import "fmt"

type NextAdvanceFutureEngine struct{}

func NewNextAdvanceFutureEngine() *NextAdvanceFutureEngine {
	return &NextAdvanceFutureEngine{}
}

func (n *NextAdvanceFutureEngine) GenerateFuturePathway(childName string, grade int, hobby string) string {
	phase := "Phase 1: Foundation (Class 6-8)"
	if grade >= 9 && grade <= 10 {
		phase = "Phase 2: Stream Selection (Class 9-10)"
	} else if grade >= 11 {
		phase = "Phase 3: Career & Entrance Target (Class 11-12)"
	}

	return fmt.Sprintf("🚀 *भविष्य का रोडमैप*\n• छात्र: %s\n• स्टेज: %s\n• हॉबी ट्रैक: %s", childName, phase, hobby)
}
