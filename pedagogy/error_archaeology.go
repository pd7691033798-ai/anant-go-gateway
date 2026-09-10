package pedagogy

type ErrorType string

const (
	CalculationSlip  ErrorType = "CALCULATION_SLIP"
	ConceptConfusion ErrorType = "CONCEPT_CONFUSION"
	RushingTrap      ErrorType = "RUSHING_TRAP"
)

type MistakeDiagnostic struct {
	IdentifiedError ErrorType
	Action          string
}

func DiagnoseStudentMistake(elapsedSec int) MistakeDiagnostic {
	if elapsedSec < 5 {
		return MistakeDiagnostic{IdentifiedError: RushingTrap, Action: "10-सेकंड टाइमर लॉक सक्रिय"}
	}
	return MistakeDiagnostic{IdentifiedError: ConceptConfusion, Action: "30-सेकंड ऑडियो हिंट अनलॉक"}
}
