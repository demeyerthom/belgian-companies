package companies

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var URL_TEMPLATE = "https://kbopub.economie.fgov.be/kbopub/toonondernemingps.html?ondernemingsnummer=%s"

type CompanyPage struct {
	Version       string
	Raw           string
	DossierNumber string
	DateFetched   time.Time
	OriginalUrl   string
}

func FetchCompanyPage(client *http.Client, dossierNumber string) (page CompanyPage, err error) {
	url := fmt.Sprintf(URL_TEMPLATE, dossierNumber)

	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return page, errors.New("something went wrong")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	raw := string(body[:])

	page.Version = "1"
	page.OriginalUrl = url
	page.DateFetched = time.Now()
	page.Raw = raw

	return page, nil
}
