package publications

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/grokify/html-strip-tags-go"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var rootUrl = "http://www.ejustice.just.fgov.be"

func ParsePublicationPage(filePath string, documentPath string, withDocuments bool) (publications []Publication, err error) {
	file, err := ioutil.ReadFile(filePath)

	if err != nil {
		panic(err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
	if err != nil {
		panic(err)
	}

	doc.Find("center table td").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		publication, err := parseNode(*s)
		if err != nil {
			return
		}

		if withDocuments {
			err = DownloadFile(documentPath+publication.FileLocation, rootUrl+publication.FileLocation)
			if err != nil {
				panic(err)
			}
		}

		publications = append(publications, publication)
	})

	return publications, nil
}

func parseNode(node goquery.Selection) (publication Publication, err error) {
	companyName := node.Find("font[color=blue]")
	documentLink, _ := node.Find("a").Attr("href")

	text, _ := node.Html()
	elements := strings.Split(text, "<br/>")

	if len(elements) < 5 {
		return publication, errors.New(fmt.Sprintf("invalid number of elements: %d", len(elements)))
	}

	publication = Publication{}
	publication.CompanyName = strings.TrimSpace(companyName.Text())
	publication.FileLocation = documentLink
	publication.Address = strings.TrimSpace(elements[1])
	publication.DossierNumber = strings.TrimSpace(elements[2])
	publication.Type = strings.TrimSpace(elements[3])

	re, _ := regexp.Compile("[0-9]{4}-[0-9]{2}-[0-9]{2}")

	datesFound := re.FindAllString(elements[4], 1)

	if len(datesFound) > 0 {
		publication.DatePublication = datesFound[0]
	}

	legalForm := strip.StripTags(elements[0])
	publication.LegalForm = strings.TrimSpace(strings.Replace(legalForm, companyName.Text(), "", 1))

	return publication, nil
}

func DownloadFile(filePath string, url string) error {
	dirPath := filepath.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, os.ModePerm)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
