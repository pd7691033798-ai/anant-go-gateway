package pricing

import (
	"fmt"
	"math"
)

type InvoiceSummary struct {
	BasePrice        int
	AddonPrice       int
	GrossTotal       int
	ReferralDiscount int
	FinalPayable     int
	AllowedKids      int
	FormattedTxt     string
}

// ComputeModularBill: बेस प्लान, परीक्षा ऐड-ऑन और रेफरल डिस्काउंट की गणना
func ComputeModularBill(baseTier string, includeExamAddon bool, successfulReferrals int, isReferredNewUser bool) (InvoiceSummary, error) {
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
	grossTotal := base + addon

	// रेफरल डिस्काउंट की गणना
	discount := calculateReferralDiscount(baseTier, successfulReferrals, isReferredNewUser)

	// फाइनल देय राशि (कभी भी 0 या माइनस में नहीं जाएगी)
	finalPayable := grossTotal - discount
	if finalPayable < 0 {
		finalPayable = 0
	}

	txt := fmt.Sprintf("बेस: ₹%d + ऐड-ऑन: ₹%d | छूट: -₹%d | कुल देय: ₹%d/माह (स्वीकृत बच्चे: %d)",
		base, addon, discount, finalPayable, kids)

	return InvoiceSummary{
		BasePrice:        base,
		AddonPrice:       addon,
		GrossTotal:       grossTotal,
		ReferralDiscount: discount,
		FinalPayable:     finalPayable,
		AllowedKids:      kids,
		FormattedTxt:     txt,
	}, nil
}

// calculateReferralDiscount: नियमों के अनुसार सटीक छूट की गणना
func calculateReferralDiscount(baseTier string, successfulReferrals int, isReferredNewUser bool) int {
	totalDiscount := 0.0

	// 1. यदि नया यूजर किसी के रेफरल लिंक से आया है तो उसे ₹100 की वेलकम छूट
	if isReferredNewUser {
		totalDiscount += 100.0
	}

	// 2. अभिभावक का रेफरल रिवॉर्ड (प्रति सफल रेफरल ₹100, अधिकतम 2 रेफरल की सख्त सीमा = ₹200)
	discountPerRef := 100.0
	maxCap := 200.0

	if baseTier == "FAMILY_1099" {
		discountPerRef = 150.0
		maxCap = 300.0
	}

	parentRefDiscount := float64(successfulReferrals) * discountPerRef
	if parentRefDiscount > maxCap {
		parentRefDiscount = maxCap
	}

	totalDiscount += parentRefDiscount

	return int(math.Round(totalDiscount))
}

// ComputeFinalBill: फ्लोट वैल्यूज के साथ अलग से क्विक कैलकुलेशन के लिए
func ComputeFinalBill(baseTier string, originalPrice float64, successfulReferrals int) float64 {
	discountPerReferral := 100.0
	maxCap := 200.0

	if baseTier == "FAMILY_1099" {
		discountPerReferral = 150.0
		maxCap = 300.0
	}

	totalDiscount := float64(successfulReferrals) * discountPerReferral
	if totalDiscount > maxCap {
		totalDiscount = maxCap
	}

	finalPayable := originalPrice - totalDiscount
	if finalPayable < 0 {
		finalPayable = 0
	}
	return math.Round(finalPayable)
}
