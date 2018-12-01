package parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type PublicationParser struct {
	*Parser
	sanitizer *bluemonday.Policy
}

const (
	DossierIndex                 = 2
	AddressIndex                 = 1
	DatePublicationIndexReversed = 1
	SubjectsIndexReversed        = 2
)

var re = regexp.MustCompile(`^(?P<date>[0-9]{4}-[0-9]{1,2}-[0-9]{1,2})\s+\/\s+(?P<publication_id>[\d]{1,8})`)

func NewPublicationParser() *PublicationParser {
	sanitizer := bluemonday.UGCPolicy()
	return &PublicationParser{sanitizer: sanitizer}
}

func (p *PublicationParser) ParsePublicationPage(result []byte) (publications []*model.Publication, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(result))
	if err != nil {
		panic(err)
	}

	doc.Find("center table td").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		publication, err := p.parseNode(*s)
		if err != nil {
			return
		}

		publications = append(publications, publication)
	})

	return publications, nil
}

func (p *PublicationParser) parseNode(node goquery.Selection) (publication *model.Publication, err error) {
	publication = model.NewPublication()
	breakCount := len(node.Find("br").Nodes)

	if breakCount >= 4 && breakCount <= 5 {
		publication.CompanyName = p.parseCompanyNameFromNode(node)
		publication.FileLocation = p.parseFileLocationFromNode(node)
		publication.Raw = p.parseRawFromNode(node)

		html, _ := node.Html()
		elements := strings.Split(html, "<br/>")
		publication.DossierNumber = p.parseDossierNumberFromElements(elements)
		publication.Address = p.parseAddressFromElements(elements)
		publication.ID, publication.DatePublication, err = p.parseIDAndDatePublicationFromElements(elements)
		publication.Subjects = p.parseSubjectsFromElements(elements)
		if err != nil {
			return publication, err
		}
	} else {
		return nil, errors.New(fmt.Sprintf("invalid number of nodes: %d", breakCount))
	}

	if p.IsInvalidPublication(publication) {
		html, _ := node.Html()
		log.WithField("raw", html).Error("an invalid publication was parsed")
		return publication, errors.New("an invalid publication was parsed")
	}

	return publication, nil
}

func (p *PublicationParser) parseDossierNumberFromElements(elements []string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}

	return reg.ReplaceAllString(elements[DossierIndex], "")
}

func (p *PublicationParser) parseAddressFromElements(elements []string) string {
	return strings.TrimSpace(elements[AddressIndex])
}

func (p *PublicationParser) parseSubjectsFromElements(elements []string) (subjects []string) {
	subjectElement := elements[len(elements)-SubjectsIndexReversed]

	for _, item := range strings.Split(subjectElement, "-") {
		subjects = append(subjects, strings.TrimSpace(item))
	}

	return subjects
}

func (p *PublicationParser) parseIDAndDatePublicationFromElements(elements []string) (publicationID int32, publicationDate string, err error) {
	dateElement := elements[len(elements)-DatePublicationIndexReversed]

	match := re.FindStringSubmatch(dateElement)
	subMatchMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	if len(subMatchMap) != 2 {
		return publicationID, publicationDate, errors.New(fmt.Sprintf("found invalid number of publication dates: %d", len(subMatchMap)))
	}

	id, _ := strconv.Atoi(subMatchMap["publication_id"])

	return int32(id), subMatchMap["date"], err
}

func (p *PublicationParser) parseCompanyNameFromNode(selection goquery.Selection) string {
	return strings.TrimSpace(selection.Find("font").Text())
}

func (p *PublicationParser) parseRawFromNode(selection goquery.Selection) string {
	text, _ := selection.Html()
	return p.sanitizer.Sanitize(text)
}

func (p *PublicationParser) parseFileLocationFromNode(selection goquery.Selection) string {
	documentLink, _ := selection.Find("a").Attr("href")
	return documentLink
}

func (p *PublicationParser) IsInvalidPublication(publication *model.Publication) bool {
	if publication.DossierNumber == "" || publication.DatePublication == "" {
		return true
	}

	return false
}

func (p *PublicationParser) DownloadFile(filePath string, url string) error {
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
