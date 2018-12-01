package fetcher

import (
	"fmt"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
	"time"
)

var urlTemplate = "http://www.ejustice.just.fgov.be/cgi_tsv/tsv_l_1.pl?lang=nl&row_id=%d&pdda=%d&pddm=%d&pddj=%d&pdfa=%d&pdfm=%d&pdfj=%d&fromtab=TSV&sql=pd+between+date%%27%d-%d-%d%%27+and+date%%27%d-%d-%d%%27+"

type PublicationFetcher struct {
	*Fetcher
}

func NewPublicationFetcher(client HttpClient, sleep int) *PublicationFetcher {
	baseFetcher := &Fetcher{client: client, maxSleep: sleep}

	return &PublicationFetcher{baseFetcher}

}

func (f PublicationFetcher) buildUrl(rowId int, period time.Time) string {
	year := period.Year()
	month := period.Month()
	day := period.Day()

	return fmt.Sprintf(urlTemplate, rowId, year, month, day, year, month, day, year, month, day, year, month, day)
}

func (f *PublicationFetcher) FetchPublicationsPage(row int, day time.Time) (result *model.PublicationPage, err error) {
	url := f.buildUrl(row, day)

	body, err := f.fetchRawPage(url)

	if err != nil {
		log.WithError(err).Errorf("An error occurred while performing an http request: %s", err)
		return nil, err
	}

	raw, _ := ioutil.ReadAll(body)

	if !strings.Contains(string(raw[:]), "table") || strings.Contains(string(raw[:]), "Einde van de lijst") {
		return nil, nil
	}

	result = model.NewPublicationPage()
	result.Raw = string(raw[:])
	result.OriginalUrl = url

	return result, err
}
