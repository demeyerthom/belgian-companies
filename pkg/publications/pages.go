package publications

import (
	"errors"
	"fmt"
	"github.com/demeyerthom/belgian-companies/pkg/models"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var urlTemplate = "http://www.ejustice.just.fgov.be/cgi_tsv/tsv_l_1.pl?lang=nl&row_id=%d&pdda=%d&pddm=%d&pddj=%d&pdfa=%d&pdfm=%d&pdfj=%d&fromtab=TSV&sql=pd+between+date%%27%d-%d-%d%%27+and+date%%27%d-%d-%d%%27+"

func buildUrl(rowId int, period time.Time) string {
	year := period.Year()
	month := period.Month()
	day := period.Day()

	return fmt.Sprintf(urlTemplate, rowId, year, month, day, year, month, day, year, month, day, year, month, day)
}

func PublicationPageExists(client *http.Client, row int, day time.Time) bool {
	url := buildUrl(row, day)

	resp, err := client.Get(url)
	if err != nil {
		log.WithError(err).Errorf("An error occurred while performing an http request: %s", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, _ := ioutil.ReadAll(resp.Body)
	raw := string(body[:])

	if !strings.Contains(raw, "table") || strings.Contains(raw, "Einde van de lijst") {
		return false
	}

	return true
}

func FetchPublicationsPage(client *http.Client, row int, day time.Time) (result *models.PublicationPage, err error) {
	url := buildUrl(row, day)

	resp, err := client.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, errors.New("something went wrong")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	raw := string(body[:])

	result = models.NewPublicationPage()
	result.Raw = raw
	result.OriginalUrl = url

	log.WithField("result", result).Debugf("returning result for url: %s", url)

	return result, err
}
