package vacation

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type DecomposedPlan struct {
	StudentPhone          string `json:"student_phone"`
	TotalAssignedTasks    int    `json:"total_assigned_tasks"`
	AllocatedVacationDays int    `json:"allocated_vacation_days"`
	DailyTaskQuota        int    `json:"daily_task_quota"`
	StatusMessage         string `json:"status_message"`
}

type HomeworkDecomposerService struct {
	db  *sql.DB
	loc *time.Location
}

func NewHomeworkDecomposerService(db *sql.DB) *HomeworkDecomposerService {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	return &HomeworkDecomposerService{
		db:  db,
		loc: loc,
	}
}

// DecomposeHomework छुट्टियों के कुल होमवर्क को दैनिक छोटे कोटे में विभाजित करके डेटाबेस में सहेजता है
func (h *HomeworkDecomposerService) DecomposeHomework(ctx context.Context, phone string, totalQuestions, vacationDays int) (*DecomposedPlan, error) {
	cleanPhone := strings.TrimSpace(phone)
	if cleanPhone == "" {
		return nil, errors.New("अमान्य फ़ोन नंबर")
	}

	if vacationDays <= 0 {
		vacationDays = 35 // डिफ़ॉल्ट ग्रीष्मकालीन अवकाश अवधि
	}

	if totalQuestions <= 0 {
		totalQuestions = 35 // न्यूनतम डिफ़ॉल्ट
	}

	// दैनिक कार्य कोटा गणना (न्यूनतम 1 प्रश्न प्रति दिन)
	dailyQuota := totalQuestions / vacationDays
	if dailyQuota < 1 {
		dailyQuota = 1
	}

	if h.db != nil {
		dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		query := `
			INSERT INTO student_holiday_assignments (
				student_phone, 
				total_assigned_tasks, 
				allocated_vacation_days, 
				daily_task_quota
			)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (student_phone) 
			DO UPDATE SET 
				total_assigned_tasks = EXCLUDED.total_assigned_tasks, 
				allocated_vacation_days = EXCLUDED.allocated_vacation_days, 
				daily_task_quota = EXCLUDED.daily_task_quota, 
				updated_at = NOW()`

		_, err := h.db.ExecContext(dbCtx, query, cleanPhone, totalQuestions, vacationDays, dailyQuota)
		if err != nil {
			return nil, err
		}
	}

	return &DecomposedPlan{
		StudentPhone:          cleanPhone,
		TotalAssignedTasks:    totalQuestions,
		AllocatedVacationDays: vacationDays,
		DailyTaskQuota:        dailyQuota,
		StatusMessage:         "अवकाश कार्य सफलतापूर्वक विभाजित किया गया।",
	}, nil
}

// IsSummerBreak जाँचता है कि क्या वर्तमान में ग्रीष्मकालीन अवकाश (17 मई से 30 जून) सक्रिय है
func (h *HomeworkDecomposerService) IsSummerBreak() bool {
	now := time.Now().In(h.loc)
	month := int(now.Month())
	day := now.Day()
	return (month == 5 && day >= 17) || month == 6
}

// IsWinterBreak जाँचता है कि क्या वर्तमान में शीतकालीन अवकाश (25 दिसंबर से 5 जनवरी) सक्रिय है
func (h *HomeworkDecomposerService) IsWinterBreak() bool {
	now := time.Now().In(h.loc)
	month := int(now.Month())
	day := now.Day()
	return (month == 12 && day >= 25) || (month == 1 && day <= 5)
}

