package vacation

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ChildCustomTrack struct {
	Phone        string `json:"phone"`
	TopicName    string `json:"topic_name"`
	IsSelfChosen bool   `json:"is_self_chosen"`
}

type CustomInterestService struct {
	db *sql.DB
}

func NewCustomInterestService(db *sql.DB) *CustomInterestService {
	return &CustomInterestService{db: db}
}

// AutoSetFromChildVoice बच्चे द्वारा बोले गए या चुने गए विषय को डेटाबेस में सहेजता है
func (c *CustomInterestService) AutoSetFromChildVoice(ctx context.Context, phone string, rawChildText string) ChildCustomTrack {
	cleanPhone := strings.TrimSpace(phone)
	cleanedTopic := strings.TrimSpace(rawChildText)
	if cleanedTopic == "" {
		cleanedTopic = "रोबोटिक्स और कार इंजन"
	}

	if c.db != nil && cleanPhone != "" {
		dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		query := `UPDATE users SET custom_interest_topic = $1 WHERE phone = $2`
		_, _ = c.db.ExecContext(dbCtx, query, cleanedTopic, cleanPhone)
	}

	return ChildCustomTrack{
		Phone:        cleanPhone,
		TopicName:    cleanedTopic,
		IsSelfChosen: true,
	}
}

// BuildChildChosenPrompt AI के लिए बच्चे की रुचि पर आधारित दैनिक 15-मिनट व्यावहारिक प्रॉम्प्ट बनाता है
func (c *CustomInterestService) BuildChildChosenPrompt(studentName string, track ChildCustomTrack, currentDay, totalDays int) string {
	cleanName := strings.TrimSpace(studentName)
	if cleanName == "" {
		cleanName = "विद्यार्थी"
	}

	topic := strings.TrimSpace(track.TopicName)
	if topic == "" {
		topic = "रचनात्मक व्यावहारिक ज्ञान"
	}

	return fmt.Sprintf(
		"=== बाल-स्वरुचि लाइफ स्किल ट्रैक (CHILD-CHOSEN SELF-INTEREST TRACK) ===\n"+
			"• विद्यार्थी: %s\n"+
			"• बच्चे द्वारा स्वयं चुना गया पसंदीदा विषय: '%s'\n"+
			"• अवकाश प्रगति: दिन %d / %d\n\n"+
			"【मार्गदर्शन एवं निर्देश】\n"+
			"1. यह विषय बच्चे ने अपनी रुचि से चुना है—उसके उत्साह और जिज्ञासा की मुक्तकंठ से सराहना करें।\n"+
			"2. आज के 15-मिनट के सत्र में '%s' से संबंधित एक सरल, व्यावहारिक और हाथ से लिखकर हल करने वाला टास्क तैयार करें।\n"+
			"3. टास्क ऐसा हो जिससे बच्चे के तार्किक चिंतन और हस्तलेखन दोनों का विकास हो।\n"+
			"4. संवाद में देसी, सरल और आत्मीय भाषा का प्रयोग करें।\n"+
			"==================================================================",
		cleanName, topic, currentDay, totalDays, topic,
	)
}
