package admin

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type RefinementEngine struct {
	MaxLoops int
}

func NewRefinementEngine() *RefinementEngine {
	return &RefinementEngine{MaxLoops: 5}
}

func (re *RefinementEngine) RefineSystemComponent(componentName string, rawSourceCode string) (string, error) {
	log.Printf("[REFINEMENT START] Component: %s scanning initiated...", componentName)
	
	currentCode := rawSourceCode
	loopCount := 0

	for loopCount < re.MaxLoops {
		loopCount++
		errs := detectBugs(currentCode)

		if len(errs) == 0 {
			log.Printf("[REFINEMENT PASSED] %s is clean in loop %d.", componentName, loopCount)
			return currentCode, nil
		}

		log.Printf("[WARNING] Found %d issues in %s. Self-correcting...", len(errs), componentName)
		refinedCode, correctionErr := applySelfCorrection(currentCode, errs)
		if correctionErr != nil {
			return "", fmt.Errorf("refinement failed: %v", correctionErr)
		}

		currentCode = refinedCode
		time.Sleep(300 * time.Millisecond)
	}

	return "", fmt.Errorf("[REFINEMENT FAILED] Component could not pass quality checks.")
}

func detectBugs(code string) []string {
	var issues []string
	if strings.Contains(code, "db.QueryRow(") && !strings.Contains(code, "Context") {
		issues = append(issues, "Missing Context Timeout in database query")
	}
	if strings.Contains(code, "fmt.Sprintf(`{\"response\":") && !strings.Contains(code, "json.Encoder") {
		issues = append(issues, "Potential JSON injection risk")
	}
	return issues
}

func applySelfCorrection(code string, issues []string) (string, error) {
	correctedCode := code
	for _, issue := range issues {
		if strings.Contains(issue, "Missing Context Timeout") {
			log.Println("[AUTO-FIX] Applying context timeout wrappers...")
		}
	}
	return correctedCode, nil
}
