package publications

import "github.com/segmentio/ksuid"

//Publication describes a new publication
type Publication struct {
	ID              string `json:"id"`
	CompanyName     string `json:"company_name"`
	LegalForm       string `json:"legal_form"`
	Address         string `json:"address"`
	FileLocation    string `json:"file_location"`
	DossierNumber   string `json:"dossier_number"`
	DatePublication string `json:"date_publication"`
	Type            string `json:"type"`
	Raw             string `json:"raw"`
}

func NewPublication() (publication Publication) {
	publication = Publication{
		ID: ksuid.New().String(),
	}

	return publication
}
