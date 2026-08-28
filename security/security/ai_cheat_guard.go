package security

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type AICheatVerdict struct {
	IsSuspicious    bool    `json:"is_suspicious"`
	CheatScore      float64 `json:"cheat_score"`
	ReasonCode      string  `json:"reason_code"`
	ShouldChallenge bool    `json:"should_challenge"`
	ChallengePrompt string  `json:"challenge_prompt"`
}

type AICheatGuardService struct {
	db *sql.DB
}

func NewAICheatGuardService(db *sql.DB) *AICheatGuardService {
	return &AICheatGuardService{db: db}
}

// EvaluateSubmissionAI छात्र के सबमिशन में बाहरी AI (ChatGPT/Gemini) के उपयोग की जाँच करता है
func (g *AICheatGuardService) EvaluateSubmissionAI(ctx context.Context, phone string, grade int, ocrText string, timeTakenSeconds int) *AICheatGuardServiceVerdict {
	cleanText := strings.ToLower(strings.TrimSpace(ocrText))
	cheatScore := 0.0
	reasons := []string{}

	// 1. LLM टिपिकल बज़वर्ड्स और औपचारिक संरचना की जाँच
	llmKeywords := []string{
		"furthermore", "in conclusion", "it is important to note", 
		"निष्कर्षतः", "उल्लेखनीय है कि", "समग्र रूप से", "संक्षेप में कहें तो",
		"as an ai", "step 1:", "step 2:", "चरण 1:", "चरण 2:",
	}

	for _, kw := range llmKeywords {
		if strings.Contains(cleanText, kw) {
			cheatScore += 25.0
			reasons = append(reasons, "LLM_FORMAL_SYNTAX")
			break
		}
	}

	// 2. असामान्य गति विश्लेषण (Fast Solving vs Complex Step Count)
	// यदि 5 से अधिक स्टेप्स का कार्य 90 सेकंड से कम में सबमिट हो गया
	if len(cleanText) > 200 && timeTakenSeconds > 0 && timeTakenSeconds < 90 {
		cheatScore += 35.0
		reasons = append(reasons, "UNNATURAL_SPEED")
	}

	// 3. कक्षा के स्तर से अत्यधिक उन्नत शब्दावली (Vocabulary Complexity)
	if grade <= 8 {
		advancedWords := []string{"दृष्टिकोण", "पारिस्थितिकी", "तदनुसार", "hypotenuse theorem", "thermodynamics", "quantification"}
		for _, w := range advancedWords {
			if strings.Contains(cleanText, w) {
				cheatScore += 20.0
				reasons = append(reasons, "OVER_COMPLEX_VOCABULARY")
				break
			}
		}
	}

	isSuspicious := cheatScore >= 40.0
	shouldChallenge := cheatScore >= 50.0

	// यदि गंभीर संदेह है तो डेटाबेस में लॉग करें
	if isSuspicious && g.db != nil && phone != "" {
		dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		_, _ = g.db.ExecContext(dbCtx, `
			UPDATE users 
			SET sharing_suspicion_score = sharing_suspicion_score + 15 
			WHERE phone = $1
		`, phone)

		_, _ = g.db.ExecContext(dbCtx, `
			INSERT INTO multi_user_violations (student_phone, detected_issue, confidence_score) 
			VALUES ($1, $2, $3)
		`, phone, "AI_ASSISTED_COPY_DETECTION", cheatScore/100.0)
	}

	challengePrompt := ""
	if shouldChallenge {
		challengePrompt = "शाबाश! आपने बहुत अच्छा उत्तर लिखा है। क्या आप संक्षेप में बता सकते हैं कि आपने दूसरे चरण में यह सूत्र/तरीका क्यों चुना?"
	}

	return &AICheatVerdict{
		IsSuspicious:    isSuspicious,
		CheatScore:      cheatScore,
		ReasonCode:      strings.Join(reasons, ","),
		ShouldChallenge: shouldChallenge,
		ChallengePrompt: challengePrompt,
	}
}

type AICheatGuardServiceVerdict = AICheatVerdict
