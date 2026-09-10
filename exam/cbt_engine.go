package exam

type CBTSessionSubmission struct {
	SessionID string            `json:"session_id"`
	StudentID string            `json:"student_id"`
	TrackCode string            `json:"track_code"`
	Answers   map[string]string `json:"answers"`
}

type CBTResult struct {
	Score      int     `json:"score"`
	TotalMarks int     `json:"total_marks"`
	Percentile float64 `json:"percentile"`
}

func EvaluateCBTSession(sub CBTSessionSubmission, answerKey map[string]string) CBTResult {
	correct := 0
	total := len(answerKey)
	for qID, correctAns := range answerKey {
		if sub.Answers[qID] == correctAns {
			correct++
		}
	}
	score := correct * 4
	pct := (float64(correct) / float64(total)) * 100.0

	return CBTResult{Score: score, TotalMarks: total * 4, Percentile: pct}
}
