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

func TestPublicationParser_ParsePublicationPage(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	raw := loadPublicationFile("publications-list.html")
	page := model.NewPublicationPage()
	page.Raw = string(raw[:])
	publications, err := parser.ParsePublicationPage(page)

	assert.Nil(err)
	assert.Len(publications, 30)
}

func TestPublicationParser_ParseInvalidDatePublication(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	raw := loadPublicationFile("invalid-date-publication.html")
	page := model.NewPublicationPage()
	page.Raw = string(raw[:])
	publications, err := parser.ParsePublicationPage(page)

	assert.Nil(err)
	assert.Len(publications, 1)
}
