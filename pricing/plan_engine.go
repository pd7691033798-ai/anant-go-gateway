package pricing

import (
	"database/sql"
)

type PlanTier string

const (
	TierDemo   PlanTier = "DEMO"
	TierBasic  PlanTier = "BASIC"
	TierPro    PlanTier = "PRO"
	TierFamily PlanTier = "FAMILY"
)

type PlanLimits struct {
	MaxDailyScans    int
	AllowedTracks    int
	DailyQAQuestions int
	MonthlyPrice     int
	MaxChildren      int
	AIAccess         bool
	MultiProfile     bool
	ExamMode         bool
	SeasonalBreak    bool // समर/विंटर सत्र और राज्य-वार छुट्टियां
	LanguageSupport  bool // स्थानीय भाषा चयन समर्थन
	TrialDays        int
}

type PlanService struct {
	db *sql.DB
}

func NewPlanService(db *sql.DB) *PlanService {
	return &PlanService{db: db}
}

func (p *PlanService) GetPlanLimits(tier PlanTier) PlanLimits {
	switch tier {
	case TierDemo:
		return PlanLimits{
			MaxDailyScans:    2,
			AllowedTracks:    1,
			DailyQAQuestions: 1,
			MonthlyPrice:     0,
			MaxChildren:      1,
			AIAccess:         false,
			MultiProfile:     false,
			ExamMode:         false,
			SeasonalBreak:    false,
			LanguageSupport:  true,
			TrialDays:        7, // 7 दिन का फ्री डेमो
		}
	case TierBasic:
		return PlanLimits{
			MaxDailyScans:    5,
			AllowedTracks:    1,
			DailyQAQuestions: 2, // बेसिक के लिए केवल 2 सवाल
			MonthlyPrice:     399,
			MaxChildren:      1,
			AIAccess:         false,
			MultiProfile:     false,
			ExamMode:         false,
			SeasonalBreak:    false,
			LanguageSupport:  true,
			TrialDays:        0,
		}
	case TierPro:
		return PlanLimits{
			MaxDailyScans:    10, // संतुलित 10 स्कैन
			AllowedTracks:    5,
			DailyQAQuestions: 10, // प्रो के लिए 10 सवाल
			MonthlyPrice:     699,
			MaxChildren:      1,
			AIAccess:         true,
			MultiProfile:     false,
			ExamMode:         true,  // परीक्षा मोड शामिल
			SeasonalBreak:    true,  // समर/विंटर और छुट्टियां शामिल
			LanguageSupport:  true,
			TrialDays:        0,
		}
	case TierFamily:
		return PlanLimits{
			MaxDailyScans:    20, // तीनों बच्चों के लिए साझा 20 स्कैन
			AllowedTracks:    10,
			DailyQAQuestions: 20, // तीनों के लिए कुल 20 सवाल
			MonthlyPrice:     899,
			MaxChildren:      3,  // 2 से 3 बच्चे
			AIAccess:         true,
			MultiProfile:     true,  // मल्टी-चाइल्ड टाइम-शेयरिंग विंडो
			ExamMode:         true,  // परीक्षा मोड शामिल
			SeasonalBreak:    true,  // समर/विंटर और छुट्टियां शामिल
			LanguageSupport:  true,
			TrialDays:        0,
		}
	default:
		// डिफ़ॉल्ट रूप से बेसिक प्लान लागू होगा
		return p.GetPlanLimits(TierBasic)
	}
}
