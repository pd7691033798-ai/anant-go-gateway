package monitor

type ElasticSession struct {
	TargetDurationSec int
	RemainingSec      int
	PenaltyAddedSec   int
	IsActive          bool
}

func NewElasticSession(targetMinutes int) *ElasticSession {
	totalSec := targetMinutes * 60
	return &ElasticSession{
		TargetDurationSec: totalSec,
		RemainingSec:      totalSec,
		PenaltyAddedSec:   0,
		IsActive:          true,
	}
}

func (s *ElasticSession) ApplyIdleOrCheatPenalty(penaltySec int) {
	s.RemainingSec += penaltySec
	s.PenaltyAddedSec += penaltySec
}

func (s *ElasticSession) TickSecond() bool {
	if !s.IsActive || s.RemainingSec <= 0 {
		return false
	}
	s.RemainingSec--
	return s.RemainingSec == 0
}

