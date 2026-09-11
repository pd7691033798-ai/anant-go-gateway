package pricing

import (
	"database/sql"
)

type PlanTier string

const (
	TierDemo            PlanTier = "DEMO"
	TierBasic           PlanTier = "BASIC"
	TierPro             PlanTier = "PRO"
	TierFamily          PlanTier = "FAMILY"
	TierUnlimitedFamily PlanTier = "UNLIMITED FAMILY"
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
			DailyQAQuestions: 3, // बेसिक के लिए केवल 3 सवाल
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
			ExamMode:         true, // परीक्षा मोड शामिल
			SeasonalBreak:    true, // समर/विंटर और छुट्टियां शामिल
			LanguageSupport:  true,
			TrialDays:        0,
		}
	case TierFamily:
		return PlanLimits{
			MaxDailyScans:    15, // तीनों बच्चों के लिए साझा 15 स्कैन
			AllowedTracks:    10,
			DailyQAQuestions: 15, // तीनों के लिए कुल 15 सवाल
			MonthlyPrice:     1099,
			MaxChildren:      3,
			AIAccess:         true,
			MultiProfile:     true, // मल्टी-चाइल्ड टाइम-शेयरिंग विंडो
			ExamMode:         true, // परीक्षा मोड शामिल
			SeasonalBreak:    true, // समर/विंटर और छुट्टियां शामिल
			LanguageSupport:  true,
			TrialDays:        0,
		}
	case TierUnlimitedFamily:
		return PlanLimits{
			MaxDailyScans:    20, // बच्चों के लिए साझा 20 स्कैन
			AllowedTracks:    10,
			DailyQAQuestions: 20, // कुल 20 सवाल
			MonthlyPrice:     1499,
			MaxChildren:      4,
			AIAccess:         true,
			MultiProfile:     true, // मल्टी-चाइल्ड टाइम-शेयरिंग विंडो
			ExamMode:         true, // परीक्षा मोड शामिल
			SeasonalBreak:    true, // समर/विंटर और छुट्टियां शामिल
			LanguageSupport:  true,
			TrialDays:        0,
		}
	default:
		// डिफ़ॉल्ट रूप से बेसिक प्लान लागू होगा
		return p.GetPlanLimits(TierBasic)
	}
}

// GetUserPlanLimits: यूजर के फ़ोन नंबर के आधार पर उसके प्लान की लिमिट्स लौटाता है
func (p *PlanService) GetUserPlanLimits(whatsappNumber string) (PlanLimits, error) {
	if p.db == nil {
		return p.GetPlanLimits(TierBasic), nil
	}

	var tierStr string
	query := `SELECT plan_tier FROM subscriptions WHERE whatsapp_number = $1 AND status = 'ACTIVE' ORDER BY id DESC LIMIT 1`
	err := p.db.QueryRow(query, whatsappNumber).Scan(&tierStr)
	if err != nil {
		// यदि कोई एक्टिव प्लान न मिले, तो डिफ़ॉल्ट बेसिक प्लान दें
		return p.GetPlanLimits(TierBasic), nil
	}

	return p.GetPlanLimits(PlanTier(tierStr)), nil
}
