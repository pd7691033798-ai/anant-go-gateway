package pricing

import "fmt"

func GetStrictChildLimit(planTier string) int {
	switch planTier {
	case "BASIC_399", "PRO_699":
		return 1 // बेसिक: 1, प्रो: 1
	case "FAMILY_1099":
		return 2 // फैमिली: सख्त 2 बच्चे
	case "UNLIMITED_1499":
		return 4 // अनलिमिटेड फैमिली: सख्त 4 बच्चे
	default:
		return 1
	}
}

func EnforceChildCap(currentCount int, planTier string) error {
	limit := GetStrictChildLimit(planTier)
	if currentCount >= limit {
		return fmt.Errorf("सीमा पूरी: %s प्लान में केवल %d बच्चे अनुमत हैं", planTier, limit)
	}
	return nil
}
