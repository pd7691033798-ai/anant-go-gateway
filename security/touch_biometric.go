package security

import (
	"errors"
	"math"
)

type TouchSample struct {
	TouchAreaRadius float64
	TouchDurationMs int64
	SwipeVelocity   float64
}

type ProxyBiometricGuard struct {
	ExpectedAgeGroup string
	ProxyStrikes     int
}

func NewProxyBiometricGuard(ageGroup string) *ProxyBiometricGuard {
	return &ProxyBiometricGuard{
		ExpectedAgeGroup: ageGroup,
		ProxyStrikes:     0,
	}
}

func (g *ProxyBiometricGuard) ValidateTouchProfile(sample TouchSample) (bool, error) {
	if g.ExpectedAgeGroup == "CHILD_5_TO_10" {
		if sample.TouchAreaRadius > 6.5 { // बड़े इंसान की चौड़ी उंगली
			g.ProxyStrikes++
			if g.ProxyStrikes >= 3 {
				return false, errors.New("PROXIMITY_FRAUD: संदिग्ध वयस्क उपस्थिति दर्ज हुई")
			}
		}
		if sample.TouchDurationMs < 40 && math.Abs(sample.SwipeVelocity) > 800 {
			g.ProxyStrikes++
		}
	}
	return true, nil
}
