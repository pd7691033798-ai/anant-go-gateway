package exam

type TrackMeta struct {
	Code        string
	DisplayName string
	TargetGrade string
	Modules     []string
}

func GetSupportedTracks() map[string]TrackMeta {
	return map[string]TrackMeta{
		"NAVODAYA":      {Code: "NAVODAYA", DisplayName: "नवोदय विद्यालय (JNVST)", TargetGrade: "5th-6th", Modules: []string{"मानसिक योग्यता", "अंकगणित", "भाषा"}},
		"SAINIK_SCHOOL": {Code: "SAINIK_SCHOOL", DisplayName: "सैनिक स्कूल (AISSEE)", TargetGrade: "6th-9th", Modules: []string{"गणित", "इंटेलिजेंस", "जीके"}},
		"NDA":           {Code: "NDA", DisplayName: "नेशनल डिफेंस एकेडमी (NDA/NA)", TargetGrade: "11th-12th", Modules: []string{"स्पीड मैथ", "डिफेंस इंग्लिश", "जीएटी"}},
		"IIT_JEE":       {Code: "IIT_JEE", DisplayName: "IIT-JEE Foundation", TargetGrade: "9th-12th", Modules: []string{"फिजिक्स", "मेंटल मैथ", "केमिस्ट्री"}},
	}
}
