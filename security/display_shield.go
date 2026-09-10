package security

import "errors"

type DisplayHardwareState struct {
	IsVirtualDisplayActive bool // AnyDesk, TeamViewer, HDMI
	ScreenCaptureDetected  bool // स्क्रीन रिकॉर्डर
	ActiveDisplaysCount    int  // 1 से अधिक डिस्प्ले
}

type DisplayShield struct {
	HardwareTamperCount int
}

func NewDisplayShield() *DisplayShield {
	return &DisplayShield{HardwareTamperCount: 0}
}

func (s *DisplayShield) InterceptScreenMirroring(state DisplayHardwareState) error {
	if state.IsVirtualDisplayActive || state.ScreenCaptureDetected || state.ActiveDisplaysCount > 1 {
		s.HardwareTamperCount++
		return errors.New("HARDWARE_TAMPER_DETECTED: स्क्रीन रिकॉर्डिंग या मिररिंग बंद करें")
	}
	return nil
}
