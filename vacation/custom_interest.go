package vacation

import (
	"database/sql"
	"fmt"
	"strings"
)

type ChildCustomTrack struct {
	Phone        string
	TopicName    string
	IsSelfChosen bool
}

type CustomInterestService struct {
	db *sql.DB
}

func NewCustomInterestService(db *sql.DB) *CustomInterestService {
	return &CustomInterestService{db: db}
}

func (c *CustomInterestService) AutoSetFromChildVoice(phone string, rawChildText string) ChildCustomTrack {
	cleanedTopic := strings.TrimSpace(rawChildText)
	if cleanedTopic == "" {
		cleanedTopic = "सामान्य रचनात्मक अभ्यास"
	}

	query := `UPDATE users SET custom_interest_topic = $1 WHERE phone = $2`
	_, _ = c.db.Exec(query, cleanedTopic, phone)

	return ChildCustomTrack{
		Phone:        phone,
		TopicName:    cleanedTopic,
		IsSelfChosen: true,
	}
}

func (c *CustomInterestService) BuildChildChosenPrompt(studentName string, track ChildCustomTrack, currentDay, totalDays int) string {
	return fmt.Sprintf(
		"=== बाल-स्वरुचि लाइफ स्किल ट्रैक (CHILD-CHOSEN) ===\n"+
		"विद्यार्थी: %s | बच्चे द्वारा खुद चुना गया विषय: '%s' | दिन %d/%d\n"+
		"निर्देश:\n"+
		"1. यह विषय बच्चे ने खुद चुना है।\n"+
		"2. विषय चाहे जो भी हो, उसे आज के 15-मिनट के हाथ से लिखने वाले व्यावहारिक अभ्यास में बदलें।\n"+
		"3. बच्चे के उत्साह की सराहना करें।",
		studentName, track.TopicName, currentDay, totalDays,
	)
}
