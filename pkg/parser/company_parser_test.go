package parser

import (
	"bytes"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

const (
	companiesAssetsDir = "../../test/data/company-pages/"
)

func loadCompanyFile(fileName string) []byte {
	page, err := ioutil.ReadFile(companiesAssetsDir + fileName)
	if err != nil {
		panic(err)
	}

	return bytes.NewBuffer(page).Bytes()
}

func TestCompanyParser_ParseCompanyPages(t *testing.T) {
	assert := assert.New(t)

	parser := NewCompanyParser()

	pages := model.NewCompanyPages()

	mainPage := model.NewCompanyPage()
	mainPage.Raw = string(loadCompanyFile("684446351-main.html")[:])
	pages.Company = mainPage

	establishmentPage := model.NewCompanyPage()
	establishmentPage.Raw = string(loadCompanyFile("684446351-establishment-2269751597.html")[:])

	pages.Establishments = append(pages.Establishments, establishmentPage)

	company, err := parser.ParseCompanyPages(pages)

	assert.Nil(err)
	assert.NotNil(company)
}
