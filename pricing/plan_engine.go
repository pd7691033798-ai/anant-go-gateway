package pricing

import (
	"context"
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
	Tier             PlanTier `json:"tier"`
	MaxDailyScans    int      `json:"max_daily_scans"`
	AllowedTracks    int      `json:"allowed_tracks"`
	DailyQAQuestions int      `json:"daily_qa_questions"`
	MonthlyPrice     int      `json:"monthly_price"`
	MaxChildren      int      `json:"max_children"`
	AIAccess         bool     `json:"ai_access"`
	MultiProfile     bool     `json:"multi_profile"`
	ExamMode         bool     `json:"exam_mode"`
	SeasonalBreak    bool     `json:"seasonal_break"`
	LanguageSupport  bool     `json:"language_support"`
	TrialDays        int      `json:"trial_days"`
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
			Tier:             TierDemo,
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
			TrialDays:        7,
		}
	case TierBasic:
		return PlanLimits{
			Tier:             TierBasic,
			MaxDailyScans:    5,
			AllowedTracks:    1,
			DailyQAQuestions: 3,
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
			Tier:             TierPro,
			MaxDailyScans:    10,
			AllowedTracks:    5,
			DailyQAQuestions: 10,
			MonthlyPrice:     699,
			MaxChildren:      1,
			AIAccess:         true,
			MultiProfile:     false,
			ExamMode:         true,
			SeasonalBreak:    true,
			LanguageSupport:  true,
			TrialDays:        0,
		}
	case TierFamily:
		return PlanLimits{
			Tier:             TierFamily,
			MaxDailyScans:    15,
			AllowedTracks:    10,
			DailyQAQuestions: 15,
			MonthlyPrice:     1099,
			MaxChildren:      3,
			AIAccess:         true,
			MultiProfile:     true,
			ExamMode:         true,
			SeasonalBreak:    true,
			LanguageSupport:  true,
			TrialDays:        0,
		}
	case TierUnlimitedFamily:
		return PlanLimits{
			Tier:             TierUnlimitedFamily,
			MaxDailyScans:    20,
			AllowedTracks:    10,
			DailyQAQuestions: 20,
			MonthlyPrice:     1499,
			MaxChildren:      4,
			AIAccess:         true,
			MultiProfile:     true,
			ExamMode:         true,
			SeasonalBreak:    true,
			LanguageSupport:  true,
			TrialDays:        0,
		}
	default:
		return p.GetPlanLimits(TierBasic)
	}
}

// GetUserPlanLimits: context.Context के साथ सुरक्षित डेटाबेस क्वेरी
func (p *PlanService) GetUserPlanLimits(ctx context.Context, userIdentifier string) (PlanLimits, error) {
	if p.db == nil {
		return p.GetPlanLimits(TierBasic), nil
	}

	var tierStr string
	query := `
		SELECT plan_tier 
		FROM subscriptions 
		WHERE (whatsapp_number = $1 OR parent_uid = $1) AND status = 'ACTIVE' 
		ORDER BY id DESC 
		LIMIT 1`

	err := p.db.QueryRowContext(ctx, query, userIdentifier).Scan(&tierStr)
	if err != nil {
		return p.GetPlanLimits(TierBasic), nil
	}

	return p.GetPlanLimits(PlanTier(tierStr)), nil
}
