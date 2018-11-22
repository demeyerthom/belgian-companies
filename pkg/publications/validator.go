package publications

import "github.com/demeyerthom/belgian-companies/pkg/models"

func IsInvalidPublication(publication *models.Publication) bool {
	if publication.DossierNumber == "" || publication.DatePublication == "" {
		return true
	}

	return false
}
