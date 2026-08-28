package pricing

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ChildProfile फैमिली या अनलिमिटेड फैमिली प्लान के तहत बच्चे का विवरण
type ChildProfile struct {
	ID         string    `json:"id"`
	ParentUID  string    `json:"parent_uid"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Grade      int       `json:"grade"`
	SchoolName string    `json:"school_name"`
	CreatedAt  time.Time `json:"created_at"`
	LockedTill time.Time `json:"locked_till"`
}

// ParentAccount मुख्य अभिभावक का विवरण
type ParentAccount struct {
	ParentUID      string    `json:"parent_uid"`
	ParentName     string    `json:"parent_name"`
	PrimaryPhone   string    `json:"primary_phone"` // WhatsApp OTP और रिपोर्ट्स इसी पर जाएंगी
	FamilySurname  string    `json:"family_surname"`
	ActiveDeviceID string    `json:"active_device_id"`
	LastActiveAt   time.Time `json:"last_active_at"`
	PlanTier       PlanTier  `json:"plan_tier"`
}

type FamilyGuardianService struct {
	db          *sql.DB
	planService *PlanService
}

func NewFamilyGuardianService(db *sql.DB, planService *PlanService) *FamilyGuardianService {
	return &FamilyGuardianService{
		db:          db,
		planService: planService,
	}
}

// 1. AddChildWithLock: दोनों प्लान्स (Family: 3 बच्चे, Unlimited: 4 बच्चे) के अनुसार जांच कर 60-दिन का लॉक लगाएगा
func (s *FamilyGuardianService) AddChildWithLock(ctx context.Context, parentUID, firstName, lastName, schoolName string, grade int) (*ChildProfile, error) {
	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" || strings.TrimSpace(schoolName) == "" {
		return nil, errors.New("बच्चे का नाम, उपनाम (Surname) और स्कूल का नाम अनिवार्य है")
	}

	// यूज़र का वर्तमान प्लान निकालें (Family या Family Unlimited)
	limits, err := s.planService.GetUserPlanLimits(ctx, parentUID)
	if err != nil {
		return nil, fmt.Errorf("प्लान लिमिट चेक करने में त्रुटि: %w", err)
	}

	if !limits.MultiProfile {
		return nil, errors.New("वर्तमान प्लान में मल्टी-चाइल्ड प्रोफाइल की अनुमति नहीं है। कृपया फैमिली प्लान में अपग्रेड करें")
	}

	// वर्तमान में जुड़े बच्चों की संख्या जांचें
	var currentCount int
	countQuery := `SELECT COUNT(*) FROM family_children WHERE parent_uid = $1`
	if err := s.db.QueryRowContext(ctx, countQuery, parentUID).Scan(&currentCount); err != nil {
		return nil, err
	}

	if currentCount >= limits.MaxChildren {
		return nil, fmt.Errorf("प्लान सीमा समाप्त: आपके %s प्लान में अधिकतम %d बच्चे ही जुड़ सकते हैं। अधिक बच्चों के लिए नया पैक लें", limits.Tier, limits.MaxChildren)
	}

	childID := generateSecureID("chld_")
	now := time.Now().UTC()
	lockDuration := 60 * 24 * time.Hour // 60 दिनों का सख़्त प्रोफ़ाइल लॉक

	insertQuery := `
		INSERT INTO family_children (id, parent_uid, first_name, last_name, grade, school_name, created_at, locked_till)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = s.db.ExecContext(ctx, insertQuery, childID, parentUID, firstName, lastName, grade, schoolName, now, now.Add(lockDuration))
	if err != nil {
		return nil, fmt.Errorf("प्रोफ़ाइल जोड़ने में त्रुटि: %w", err)
	}

	return &ChildProfile{
		ID:         childID,
		ParentUID:  parentUID,
		FirstName:  firstName,
		LastName:   lastName,
		Grade:      grade,
		SchoolName: schoolName,
		CreatedAt:  now,
		LockedTill: now.Add(lockDuration),
	}, nil
}

// 2. ValidateProfileModification: 60 दिनों से पहले किसी भी बदलाव या स्लॉट खाली करने को रोकेगा
func (s *FamilyGuardianService) ValidateProfileModification(ctx context.Context, childID string) error {
	var lockedTill time.Time
	query := `SELECT locked_till FROM family_children WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, childID).Scan(&lockedTill)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("प्रोफ़ाइल नहीं मिली")
		}
		return err
	}

	if time.Now().UTC().Before(lockedTill) {
		remainingDays := int(time.Until(lockedTill).Hours() / 24)
		return fmt.Errorf("यह प्रोफ़ाइल अगले %d दिनों के लिए सुरक्षित लॉक है। इसे बदला या हटाया नहीं जा सकता", remainingDays)
	}

	return nil
}

// 3. EnforceSingleActiveDevice: समवर्ती (Concurrent) सेशन को रोककर पिछले डिवाइस को लॉगआउट करेगा
func (s *FamilyGuardianService) EnforceSingleActiveDevice(ctx context.Context, parentUID, currentDeviceID string) error {
	if currentDeviceID == "" {
		return errors.New("अमान्य डिवाइस आईडी")
	}

	now := time.Now().UTC()
	query := `
		UPDATE parent_accounts 
		SET active_device_id = $1, last_active_at = $2 
		WHERE parent_uid = $3
	`
	res, err := s.db.ExecContext(ctx, query, currentDeviceID, now, parentUID)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("अभिभावक खाता नहीं मिला")
	}

	return nil
}

// 4. ValidateDeviceSession: पुष्टि करता है कि केवल अधिकृत सक्रिय डिवाइस ही चल रहा है
func (s *FamilyGuardianService) ValidateDeviceSession(ctx context.Context, parentUID, requestingDeviceID string) (bool, error) {
	var activeDeviceID string
	query := `SELECT COALESCE(active_device_id, '') FROM parent_accounts WHERE parent_uid = $1`
	err := s.db.QueryRowContext(ctx, query, parentUID).Scan(&activeDeviceID)
	if err != nil {
		return false, err
	}

	if activeDeviceID != requestingDeviceID {
		return false, errors.New("सेशन समाप्त: आपका खाता किसी अन्य डिवाइस पर सक्रिय किया गया है")
	}

	return true, nil
}

// 5. GenerateReportCardHeader: सर्टिफिकेट और रिपोर्ट पर मुख्य अभिभावक का नाम मुद्रित करता है
func (s *FamilyGuardianService) GenerateReportCardHeader(ctx context.Context, childID string) (string, error) {
	var childFirst, childLast, parentName, schoolName string
	query := `
		SELECT c.first_name, c.last_name, c.school_name, p.parent_name 
		FROM family_children c
		JOIN parent_accounts p ON c.parent_uid = p.parent_uid
		WHERE c.id = $1
	`
	err := s.db.QueryRowContext(ctx, query, childID).Scan(&childFirst, &childLast, &schoolName, &parentName)
	if err != nil {
		return "", err
	}

	header := fmt.Sprintf("विद्यार्थी: %s %s | स्कूल: %s | अभिभावक: %s (पंजीकृत खाताधारक)", childFirst, childLast, schoolName, parentName)
	return header, nil
}

func generateSecureID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}
