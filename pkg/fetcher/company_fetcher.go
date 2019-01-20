package fetcher

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
)

const (
	rootPageTemplate              = "https://kbopub.economie.fgov.be/kbopub/"
	companyMainPageTemplate       = rootPageTemplate + "zoeknummerform.html?lang=en&nummer=%s"
	establishmentPageListTemplate = rootPageTemplate + "vestiginglijst.html?lang=en&ondernemingsnummer=%s&page=%d"

	establishmentsPerPage = 20
)

type CompanyFetcher struct {
	*Fetcher
}

func NewCompanyFetcher(client HttpClient, sleep int) *CompanyFetcher {
	baseFetcher := &Fetcher{client: client, maxSleep: sleep}

	return &CompanyFetcher{baseFetcher}
}

func (f *CompanyFetcher) CompanyPageExists(dossierNumber string) bool {
	url := fmt.Sprintf(companyMainPageTemplate, dossierNumber)

	resp, err := f.client.Get(url)
	if err != nil {
		log.WithError(err).Errorf("An error occurred while performing an http request: %s", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}

func (f *CompanyFetcher) FetchCompanyPages(dossierNumber string) (result *model.CompanyPages, err error) {
	url := fmt.Sprintf(companyMainPageTemplate, dossierNumber)
	body, err := f.fetchRawPage(url)
	if err != nil {
		return nil, err
	}

	result = model.NewCompanyPages()
	var pages []*model.CompanyPage

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	establishmentNo := 0
	doc.Find("table tr td.QL").Each(func(i int, selection *goquery.Selection) {
		html, _ := selection.Html()
		if !strings.Contains(html, "Number of establishment units") {
			return
		}

		establishmentNodes := selection.Parent().Find("td strong")
		if establishmentNodes.Length() != 1 {
			nodeHtml, _ := establishmentNodes.Html()
			log.WithField("html", nodeHtml).Warningf("establishment nodes found with invalid node length: %d", establishmentNodes.Length())
			return
		}

		count, _ := establishmentNodes.Html()
		intCount, _ := strconv.Atoi(count)
		establishmentNo = intCount
	})

	if establishmentNo != 0 {
		establishmentPages, err := f.fetchEstablishmentListPages(dossierNumber, establishmentNo)
		if err != nil {
			return nil, err
		}

		pages = append(pages, establishmentPages...)
	}

	raw, _ := ioutil.ReadAll(body)
	companyPage := model.NewCompanyPage()
	companyPage.Raw = string(raw[:])
	companyPage.OriginalUrl = url
	result.Company = companyPage

	result.Establishments = pages

	return result, err
}

func (f *CompanyFetcher) fetchEstablishmentListPages(dossierNumber string, establishmentNo int) (pages []*model.CompanyPage, err error) {
	pageCount := int(math.Ceil(float64(establishmentNo) / float64(establishmentsPerPage)))

	for i := 1; i <= pageCount; i++ {
		fetchedPages, err := f.fetchEstablishmentsPages(dossierNumber, i)
		if err != nil {
			return pages, err
		}

		pages = append(pages, fetchedPages...)
	}

	return pages, err
}

func (f *CompanyFetcher) fetchEstablishmentsPages(dossierNumber string, pageNumber int) (pages []*model.CompanyPage, err error) {
	url := fmt.Sprintf(establishmentPageListTemplate, dossierNumber, pageNumber)
	body, err := f.fetchRawPage(url)
	if err != nil {
		return pages, err
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return pages, err
	}

	doc.Find("table#vestiginglist tbody tr a").Each(func(key int, selection *goquery.Selection) {
		val, exists := selection.Attr("href")
		if !exists {
			return
		}

		url := rootPageTemplate + val + "&lang=en"
		body, err := f.fetchRawPage(url)
		if err != nil {
			return
		}

		raw, _ := ioutil.ReadAll(body)

		page := model.NewCompanyPage()
		page.OriginalUrl = rootPageTemplate + val
		page.Raw = string(raw[:])

		pages = append(pages, page)
	})

	return pages, nil
}
