package parser

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

const (
	assetsDir = "../../test/data/publication-pages/"
)

func loadFile(fileName string) []byte {
	page, err := ioutil.ReadFile(assetsDir + fileName)
	if err != nil {
		panic(err)
	}

	return bytes.NewBuffer(page).Bytes()
}

func TestPublicationParser_ParsePublicationPage(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	publications, err := parser.ParsePublicationPage(loadFile("publications-list.html"))

	assert.Nil(err)
	assert.Len(publications, 30)
}

func TestPublicationParser_ParseInvalidDatePublication(t *testing.T) {
	assert := assert.New(t)

	parser := NewPublicationParser()

	publications, err := parser.ParsePublicationPage(loadFile("invalid-date-publication.html"))

	assert.Nil(err)
	assert.Len(publications, 1)
}
