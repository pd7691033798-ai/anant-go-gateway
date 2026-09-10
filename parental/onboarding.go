package parental

type OnboardingStep struct {
	StepID      string `json:"step_id"`
	Title       string `json:"title"`
	ActionType  string `json:"action_type"`
	TimeoutSec  int    `json:"timeout_sec"`
}

func GetParentWalkthrough() []OnboardingStep {
	return []OnboardingStep{
		{StepID: "LOGIC_DEMO", Title: "15-सेकंड नवोदय पैटर्न डेमो", ActionType: "SOLVE_PUZZLE", TimeoutSec: 15},
		{StepID: "CHEAT_FREEZE_DEMO", Title: "नो-चीटिंग लाइव टाइमर फ्रीज डेमो", ActionType: "TEST_FREEZE", TimeoutSec: 15},
		{StepID: "MASTER_PIN_SETUP", Title: "अभिभावक 4-अंकीय मास्टर पिन लॉक", ActionType: "PIN_SETUP", TimeoutSec: 30},
	}
}
