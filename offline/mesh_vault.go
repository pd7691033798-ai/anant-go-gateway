package offline

import (
	"crypto/sha256"
	"fmt"
)

type LocalVault struct {
	StudentID         string
	DaysPreloaded     int
	LastMonotonicTick int64
}

func NewLocalVault(studentID string, days int) *LocalVault {
	return &LocalVault{StudentID: studentID, DaysPreloaded: days, LastMonotonicTick: 0}
}

func (v *LocalVault) AccessDailySprint(currentTick int64, dayIndex int) (bool, error) {
	if currentTick < v.LastMonotonicTick {
		return false, fmt.Errorf("समय से छेड़छाड़ पकड़ी गई")
	}
	if dayIndex > v.DaysPreloaded {
		return false, fmt.Errorf("30-दिवसीय ऑफ़लाइन कोटा समाप्त")
	}
	v.LastMonotonicTick = currentTick
	return true, nil
}

func (v *LocalVault) GenerateDeltaSyncPayload(completedCount int) string {
	raw := fmt.Sprintf("ID:%s|SYNC_DAYS:%d", v.StudentID, completedCount)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%s|CHECKSUM:%x", raw, hash[:4])
}
