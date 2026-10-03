package features

import (
	"database/sql"
	"strings"
)

type FeatureSet struct {
	MaxChildren   int
	ExamMode      bool
	HobbyCamp     bool
	RegionalLangs bool
	GlobalAccent  bool
	OlympiadPrep  bool
}

var PlanPermissions = map[string]FeatureSet{
	"BASIC":     {MaxChildren: 1, ExamMode: false, HobbyCamp: false, RegionalLangs: true},
	"PRO":       {MaxChildren: 1, ExamMode: true, HobbyCamp: true, RegionalLangs: true},
	"FAMILY":    {MaxChildren: 2, ExamMode: true, HobbyCamp: true, RegionalLangs: true},
	"UNLIMITED": {MaxChildren: 3, ExamMode: true, HobbyCamp: true, RegionalLangs: true},
}

func CanUserAccessFeature(db *sql.DB, planName string, featureKey string) bool {
	planName = strings.ToUpper(planName)
	var allowed bool
	query := `SELECT is_allowed FROM plan_features WHERE plan_name = $1 AND feature_key = $2`
	err := db.QueryRow(query, planName, strings.ToUpper(featureKey)).Scan(&allowed)
	if err == nil {
		return allowed
	}

	features, exists := PlanPermissions[planName]
	if !exists {
		return false
	}

	switch featureKey {
	case "EXAM_MODE":
		return features.ExamMode
	case "HOBBY_CAMP":
		return features.HobbyCamp
	default:
		return false
	}
}
