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
	State                 string `json:"state"`
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
func (h *HomeworkDecomposerService) DecomposeHomework(ctx context.Context, phone, state string, totalQuestions, vacationDays int) (*DecomposedPlan, error) {
	cleanPhone := strings.TrimSpace(phone)
	if cleanPhone == "" {
		return nil, errors.New("अमान्य फ़ोन नंबर")
	}

	cleanState := strings.TrimSpace(state)
	if cleanState == "" {
		cleanState = "Rajasthan"
	}

	if vacationDays <= 0 {
		vacationDays = 35 // डिफ़ॉल्ट अवकाश अवधि
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
		State:                 cleanState,
		TotalAssignedTasks:    totalQuestions,
		AllocatedVacationDays: vacationDays,
		DailyTaskQuota:        dailyQuota,
		StatusMessage:         "अवकाश कार्य सफलतापूर्वक विभाजित किया गया।",
	}, nil
}

// IsSummerBreak राज्य और क्षेत्रीय कैलेंडर के अनुसार ग्रीष्मकालीन अवकाश की पुष्टि करता है
func (h *HomeworkDecomposerService) IsSummerBreak(state string) bool {
	now := time.Now().In(h.loc)
	month := int(now.Month())
	day := now.Day()
	cleanState := strings.TrimSpace(state)

	switch cleanState {
	case "Jammu and Kashmir", "Ladakh", "Himachal Pradesh":
		// पहाड़ी राज्यों में गर्मियों की छुट्टियां छोटी (जुलाई में 1-2 हफ्ते) होती हैं
		return month == 7 && (day >= 1 && day <= 15)

	case "Maharashtra", "Gujarat", "Karnataka", "Tamil Nadu", "Kerala", "Andhra Pradesh", "Telangana":
		// पश्चिम एवं दक्षिण भारत: 15 अप्रैल से मई के अंत तक
		return (month == 4 && day >= 15) || month == 5

	case "Rajasthan", "Haryana", "Punjab", "Delhi", "Uttar Pradesh", "Bihar", "Madhya Pradesh", "Uttarakhand":
		// उत्तर भारत के मैदानी क्षेत्र: 17 मई से 30 जून
		return (month == 5 && day >= 17) || month == 6

	default:
		// सामान्य डिफ़ॉल्ट
		return (month == 5 && day >= 17) || month == 6
	}
}

// IsWinterBreak राज्य और जलवायु के अनुसार सटीक पैन-इंडिया शीतकालीन अवकाश की जाँच करता है
func (h *HomeworkDecomposerService) IsWinterBreak(state string) bool {
	now := time.Now().In(h.loc)
	month := int(now.Month())
	day := now.Day()
	cleanState := strings.TrimSpace(state)

	switch cleanState {
	case "Jammu and Kashmir", "Ladakh", "Himachal Pradesh":
		// पहाड़ी क्षेत्र: भारी बर्फबारी के कारण लंबा ब्रेक (15 दिसंबर से 28 फरवरी)
		return (month == 12 && day >= 15) || month == 1 || month == 2

	case "Rajasthan", "Haryana", "Punjab", "Delhi", "Uttar Pradesh", "Bihar", "Madhya Pradesh", "Uttarakhand":
		// उत्तर भारत: कड़ाके की ठंड (25 दिसंबर से 15 जनवरी)
		return (month == 12 && day >= 25) || (month == 1 && day <= 15)

	case "Tamil Nadu", "Andhra Pradesh", "Telangana", "Karnataka":
		// दक्षिण भारत: संक्रांति / पोंगल अवकाश (10 जनवरी से 18 जनवरी)
		return month == 1 && (day >= 10 && day <= 18)

	case "Maharashtra", "Gujarat":
		// पश्चिम भारत: क्रिसमस / शीतकालीन ब्रेक (24 दिसंबर से 1 जनवरी)
		return (month == 12 && day >= 24) || (month == 1 && day <= 1)

	default:
		// सामान्य डिफ़ॉल्ट
		return (month == 12 && day >= 25) || (month == 1 && day <= 5)
	}
}
