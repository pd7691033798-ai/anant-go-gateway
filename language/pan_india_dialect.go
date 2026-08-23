package language

type RegionalProfile struct {
	State       string
	District    string
	DialectName string
	ToneHint    string
}

type PanIndiaDialectService struct{}

func NewPanIndiaDialectService() *PanIndiaDialectService {
	return &PanIndiaDialectService{}
}

func (p *PanIndiaDialectService) ResolveRegionalDialect(state, district string) RegionalProfile {
	if state == "Rajasthan" {
		switch district {
		case "Sri Ganganagar", "Hanumangarh", "Anupgarh":
			return RegionalProfile{State: state, District: district, DialectName: "BAGRI", ToneHint: "घणो आछो / शाबाश टाबर"}
		case "Jodhpur", "Bikaner", "Barmer", "Nagaur":
			return RegionalProfile{State: state, District: district, DialectName: "MARWARI", ToneHint: "घणी फूटरी राइटिंग / शाबाश"}
		case "Jaipur", "Dausa", "Dudu":
			return RegionalProfile{State: state, District: district, DialectName: "DHUNDHARI", ToneHint: "घणो चोखो काम / बढ़िया"}
		case "Sikar", "Jhunjhunu", "Churu":
			return RegionalProfile{State: state, District: district, DialectName: "SHEKHAWATI", ToneHint: "घणो जोरदार / शाबाश"}
		case "Kota", "Bundi", "Baran", "Jhalawar":
			return RegionalProfile{State: state, District: district, DialectName: "HADOTI", ToneHint: "घणो चोखो बेटा / शाबाश"}
		default:
			return RegionalProfile{State: state, District: district, DialectName: "STANDARD_RAJASTHANI", ToneHint: "बहुत बढ़िया / शाबाश"}
		}
	}

	switch state {
	case "Punjab":
		return RegionalProfile{State: state, District: district, DialectName: "PUNJABI", ToneHint: "ਬਹੁਤ ਵਧੀਆ / ਸ਼ਾਬਾਸ਼ ਪੁੱਤਰ"}
	case "Haryana":
		return RegionalProfile{State: state, District: district, DialectName: "HARYANVI", ToneHint: "घणा बढ़िया लिख्या / शाबाश बालक"}
	case "Uttar Pradesh", "Bihar":
		return RegionalProfile{State: state, District: district, DialectName: "HINDI_REGIONAL", ToneHint: "बहुत सुंदर लिखावट / शाबाश"}
	case "Maharashtra":
		return RegionalProfile{State: state, District: district, DialectName: "MARATHI", ToneHint: "खूप छान / उत्तम"}
	default:
		return RegionalProfile{State: state, District: district, DialectName: "STANDARD_HINDI", ToneHint: "बहुत बढ़िया / शाबाश"}
	}
}
