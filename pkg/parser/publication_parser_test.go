package parser

import (
	"bytes"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

const (
	publicationsAssetsDir = "../../test/data/publication-pages/"
)

func loadPublicationFile(fileName string) []byte {
	page, err := ioutil.ReadFile(publicationsAssetsDir + fileName)
	if err != nil {
		panic(err)
	}

	return bytes.NewBuffer(page).Bytes()
}

func TestPublicationParser_ParsePublicationPageList(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	raw := loadPublicationFile("publications-list.html")
	page := model.NewPublicationPage()
	page.Raw = string(raw[:])
	publications, err := parser.ParsePublicationPage(page)

	assert.Nil(err)
	assert.Len(publications, 30)
}

func TestPublicationParser_ParseBasic(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	raw := loadPublicationFile("basic.html")
	page := model.NewPublicationPage()
	page.Raw = string(raw[:])
	publications, err := parser.ParsePublicationPage(page)

	assert.Nil(err)
	assert.Len(publications, 1)
	publication := publications[0]

	assert.Equal(167036, publication.ID)
	assert.Equal("2018-11-23", publication.DatePublication)
	assert.Equal("650768050", publication.DossierNumber)
	assert.Equal("BVBA", publication.LegalType)
	assert.Equal("AMINI CARS", publication.CompanyName)
	assert.Equal("RUE DE MOORSLEDE 149 1020 BRUSSEL", publication.Address)
	assert.Equal("", publication.Comment)
	assert.Equal("/tsv_pdf/2018/11/23/18167036.pdf", publication.FileLocation)
	assert.Len(publication.Subjects, 2)
	assert.Equal("KAPITAAL", publication.Subjects[0])
	assert.Equal("AANDELEN", publication.Subjects[1])
}

func TestPublicationParser_ParseWithAdditionalCode(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	raw := loadPublicationFile("additional-code.html")
	page := model.NewPublicationPage()
	page.Raw = string(raw[:])
	publications, err := parser.ParsePublicationPage(page)

	assert.Nil(err)
	publication := publications[0]
	assert.Equal("DEMISSIONS, NOMINATIONS", publication.Subjects[0])
}

func TestPublicationParser_ParseAdditionalComment(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	raw := loadPublicationFile("additional-comment.html")
	page := model.NewPublicationPage()
	page.Raw = string(raw[:])
	publications, err := parser.ParsePublicationPage(page)

	assert.Nil(err)
	publication := publications[0]
	assert.Equal("Annulatie van de ambtshalve doorhaling.", publication.Comment)
}
