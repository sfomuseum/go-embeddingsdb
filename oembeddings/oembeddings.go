package oembeddings

type OEmbeddings struct {
	Type              string `json:"type"`
	Preview           string `json:"preview"`
	DepictionURL      string `json:"depiction_url"`
	SubjectURL        string `json:"subject_url"`
	SubjectTitle      string `json:"subject_title"`
	SubjectCreditline string `json:"subject_creditline"`
	ProviderName      string `json:"provider_name"`
	ProviderURL       string `json:"provider_url"`
}
