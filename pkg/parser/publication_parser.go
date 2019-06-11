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

var (
	CompanyNameIndex   = 0
	AddressIndex       = 1
	DossierNumberIndex = 2
	SubjectIndex       = 3
)

var dossierNumberRegex = regexp.MustCompile("[^a-zA-Z0-9]+")
var dateAndPublicationRegex = regexp.MustCompile(`^(?P<date>[0-9]{4}-[0-9]{1,2}-[0-9]{1,2})\s+\/\s+(?P<publication_id>[\d]{1,8})`)

func NewPublicationParser() *PublicationParser {
	sanitizer := bluemonday.UGCPolicy()
	return &PublicationParser{sanitizer: sanitizer}
}

func (p *PublicationParser) ParsePublicationPage(result *model.PublicationPage) (publications []*model.Publication, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(result.Raw)))
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

	if breakCount < 1 {
		return nil, errors.New(fmt.Sprintf("invalid number of nodes: %d", breakCount))
	}
	publication.Raw = p.returnRawFromNode(node)

	html, _ := node.Html()
	elements := strings.Split(html, "<br/>")
	publication.CompanyName, publication.LegalType = p.parseCompanyNameAndLegalFormFromElements(elements)
	publication.DossierNumber = p.parseDossierNumberFromElements(elements)
	publication.Address = p.parseAddressFromElements(elements)
	publication.ID, publication.DatePublication, publication.FileLocation = p.parseIDDatePublicationAndFileLocationFromElements(elements)
	publication.Subjects = p.parseSubjects(elements)

	if p.IsInvalidPublication(publication) {
		html, _ := node.Html()
		log.WithField("raw", html).Error("an invalid publication was parsed")
		return publication, errors.New("an invalid publication was parsed")
	}

	return publication, nil
}

func (p *PublicationParser) parseDossierNumberFromElements(elements []string) string {
	return dossierNumberRegex.ReplaceAllString(elements[DossierNumberIndex], "")
}

func (p *PublicationParser) parseAddressFromElements(elements []string) string {
	return strings.TrimSpace(elements[AddressIndex])
}

func (p *PublicationParser) parseIDDatePublicationAndFileLocationFromElements(elements []string) (publicationID int32, publicationDate string, fileLocation string) {
	var subMatchMap = make(map[string]string)
	var key int

	for i, element := range elements {
		match := dateAndPublicationRegex.FindStringSubmatch(element)

		if len(match) != 3 {
			continue
		}

		for i, name := range dateAndPublicationRegex.SubexpNames() {
			subMatchMap[name] = match[i]
		}
		key = i
		break
	}

	if len(subMatchMap) == 0 {
		return publicationID, publicationDate, fileLocation
	}

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader([]byte(elements[key])))
	fileLocation, _ = doc.Find("a").Attr("href")

	id, _ := strconv.Atoi(subMatchMap["publication_id"])

	return int32(id), subMatchMap["date"], fileLocation
}

func (p *PublicationParser) parseCompanyNameAndLegalFormFromElements(elements []string) (name string, legalForm string) {
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader([]byte(elements[CompanyNameIndex])))
	name, _ = doc.Find("font").Html()
	legalForm = strings.TrimSpace(strings.Replace(doc.Text(), name, "", 1))

	return name, legalForm
}

func (p *PublicationParser) returnRawFromNode(selection goquery.Selection) string {
	text, _ := selection.Html()
	return p.sanitizer.Sanitize(text)
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

func (p *PublicationParser) parseSubjects(elements []string) (subjects []string) {
	subjectsElement := elements[SubjectIndex]

	rawSubjects := strings.Split(subjectsElement, "-")
	for _, rawSubject := range rawSubjects {
		subjects = append(subjects, strings.TrimSpace(rawSubject))
	}

	return subjects
}
