package pricing

import "fmt"

type InvoiceSummary struct {
	BasePrice    int
	AddonPrice   int
	TotalPrice   int
	AllowedKids  int
	FormattedTxt string
}

func ComputeModularBill(baseTier string, includeExamAddon bool) (InvoiceSummary, error) {
	base := 0
	switch baseTier {
	case "BASIC_399":
		base = 399
	case "PRO_699":
		base = 699
	case "FAMILY_1099":
		base = 1099
	case "UNLIMITED_1499":
		base = 1499
	default:
		return InvoiceSummary{}, fmt.Errorf("अमान्य प्लान: %s", baseTier)
	}

	addon := 0
	if includeExamAddon {
		addon = 799
	}

	kids := GetStrictChildLimit(baseTier)
	total := base + addon

	return InvoiceSummary{
		BasePrice:    base,
		AddonPrice:   addon,
		TotalPrice:   total,
		AllowedKids:  kids,
		FormattedTxt: fmt.Sprintf("बेस: ₹%d + परीक्षा विंग: ₹%d = कुल: ₹%d/माह (बच्चे: %d)", base, addon, total, kids),
	}, nil
}
