package pricing

import "database/sql"

type PlanTier string

const (
	TierDemo  PlanTier = "DEMO"
	TierBasic PlanTier = "BASIC"
	TierPro   PlanTier = "PRO"
)

type PlanLimits struct {
	MaxDailyScans  int
	AllowedTracks  int
	CanDecomposeHW bool
	HasBiometrics  bool
	TrialDays      int
}

type PlanService struct {
	db *sql.DB
}

func NewPlanService(db *sql.DB) *PlanService {
	return &PlanService{db: db}
}

func (p *PlanService) GetPlanLimits(tier PlanTier) PlanLimits {
	switch tier {
	case TierPro:
		return PlanLimits{MaxDailyScans: 15, AllowedTracks: 5, CanDecomposeHW: true, HasBiometrics: true, TrialDays: 0}
	case TierBasic:
		return PlanLimits{MaxDailyScans: 6, AllowedTracks: 1, CanDecomposeHW: true, HasBiometrics: true, TrialDays: 0}
	default:
		return PlanLimits{MaxDailyScans: 2, AllowedTracks: 0, CanDecomposeHW: false, HasBiometrics: false, TrialDays: 7}
	}
}
