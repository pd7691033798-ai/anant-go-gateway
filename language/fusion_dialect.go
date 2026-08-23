package language

import (
	"database/sql"
	"strings"
)

type DialectProfile struct {
	DialectCode string
	ToneHint    string
	RegionHint  string
	IsFusion    bool
}

type FusionDialectService struct {
	db *sql.DB
}

func NewFusionDialectService(db *sql.DB) *FusionDialectService {
	return &FusionDialectService{db: db}
}

func (f *FusionDialectService) DetectAndResolve(rawInput, state, district string) DialectProfile {
	lower := strings.ToLower(rawInput)

	hasBagri := strings.Contains(lower, "ठा कोनी") || strings.Contains(lower, "मैंने") || strings.Contains(lower, "म्हारो") || strings.Contains(lower, "घणो")
	hasPunjabi := strings.Contains(lower, "दस") || strings.Contains(lower, "कहना चाहना") || strings.Contains(lower, "ਕੀ") || strings.Contains(lower, "ਪੁੱਤਰ") || strings.Contains(lower, "ਚੱਕ")

	if hasBagri && hasPunjabi {
		return DialectProfile{
			DialectCode: "BAGRI_PUNJABI_FUSION",
			ToneHint:    "मन्ने सब ठा है, तू की कहना चौहना ऐ! चल कॉपी चक्क ते लिख!",
			RegionHint:  "श्रीगंगानगर/फाजिल्का सीमावर्ती बेल्ट",
			IsFusion:    true,
		}
	}

	if district == "Sri Ganganagar" || district == "Hanumangarh" {
		return DialectProfile{DialectCode: "BAGRI", ToneHint: "घणो आछो लिख्यो है / शाबाश टाबर", RegionHint: "उत्तर राजस्थान", IsFusion: false}
	}

	return DialectProfile{DialectCode: "STANDARD_HINDI", ToneHint: "बहुत बढ़िया लिखा है / शाबाश", RegionHint: "अखिल भारतीय", IsFusion: false}
}
