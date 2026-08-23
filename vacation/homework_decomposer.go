package vacation

import (
	"database/sql"
	"time"
)

type HomeworkService struct {
	db *sql.DB
}

func NewHomeworkService(db *sql.DB) *HomeworkService {
	return &HomeworkService{db: db}
}

func (h *HomeworkService) DecomposeHomework(phone string, totalQuestions, vacationDays int) error {
	if vacationDays <= 0 {
		vacationDays = 35
	}
	dailyQuota := totalQuestions / vacationDays
	if dailyQuota < 1 {
		dailyQuota = 1
	}

	query := `INSERT INTO student_holiday_assignments (student_phone, total_assigned_tasks, allocated_vacation_days, daily_task_quota)
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (student_phone) 
	          DO UPDATE SET total_assigned_tasks = $2, allocated_vacation_days = $3, daily_task_quota = $4, updated_at = NOW()`
	_, err := h.db.Exec(query, phone, totalQuestions, vacationDays, dailyQuota)
	return err
}

func (h *HomeworkService) IsSummerBreak() bool {
	now := time.Now()
	month := int(now.Month())
	day := now.Day()
	return (month == 5 && day >= 17) || month == 6
}
