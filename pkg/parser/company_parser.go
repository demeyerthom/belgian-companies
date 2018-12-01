package parser

import (
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/microcosm-cc/bluemonday"
)

type CompanyParser struct {
	*Parser
	sanitizer *bluemonday.Policy
}

func NewCompanyParser() *CompanyParser {
	sanitizer := bluemonday.UGCPolicy()
	return &CompanyParser{sanitizer: sanitizer}
}

func (p *CompanyParser) ParseCompanyPages(pages *model.CompanyPages) (*model.Company, error) {
	company := model.NewCompany()

	return company, nil
}
