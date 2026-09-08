package parental

import (
	"fmt"
	"time"
)

type FocusLevel string

const (
	ModeEasy       FocusLevel = "EASY"
	ModeModerate   FocusLevel = "MODERATE"
	ModeRestricted FocusLevel = "RESTRICTED"
)

type ChildFocusSession struct {
	StudentID       string
	SelectedMode    FocusLevel
	IsAppLocked     bool
	IsCallActive    bool
	SessionPausedAt time.Time
	RemainingSec    int
}

// HandleIncomingCall: कॉल आने पर ऐप को पॉज करना
func (c *ChildFocusSession) HandleIncomingCall() string {
	c.IsCallActive = true
	c.SessionPausedAt = time.Now()
	return "📞 [CALL PERMITTED] अभ्यास अस्थायी रूप से रुका। कॉल समाप्त होने पर वापस लॉक सक्रिय होगा।"
}

// HandleCallEnded: कॉल कटते ही रिस्ट्रिक्टेड मोड वापस चालू करना
func (c *ChildFocusSession) HandleCallEnded() string {
	c.IsCallActive = false
	c.IsAppLocked = (c.SelectedMode == ModeRestricted)
	return fmt.Sprintf("🔒 [FOCUS RESTORED] कॉल समाप्त। %s मोड पुनः सक्रिय। अभ्यास जारी रखें।", c.SelectedMode)
}

// parental/focus_mode.go के सबसे नीचे जोड़ें:

// RunFocusDemo: main.go से सीधे कॉल करने के लिए सिंगल एंट्री पॉइंट
func RunFocusDemo() {
	session := &ChildFocusSession{
		StudentID:    "STUDENT-101",
		SelectedMode: ModeRestricted,
		IsAppLocked:  true,
		RemainingSec: 900,
	}

	fmt.Printf("✅ [Parental] मोड: %s | स्क्रीन लॉक: %t\n", session.SelectedMode, session.IsAppLocked)
	fmt.Printf("   %s\n", session.HandleIncomingCall())
	fmt.Printf("   %s\n", session.HandleCallEnded())
}
