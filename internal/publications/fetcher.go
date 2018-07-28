package publications

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var urlTemplate = "http://www.ejustice.just.fgov.be/cgi_tsv/tsv_l_1.pl?lang=nl&row_id=%d&pdda=%d&pddm=%d&pddj=%d&pdfa=%d&pdfm=%d&pdfj=%d&fromtab=TSV_TMP&rech=1003&sql=pd+between+date%%27%d-%d-%d%%27+and+date%%27%d-%d-%d%%27+"

type FetchedPublicationPage struct {
	Raw         string    `bson:"raw" json:"raw"`
	DateAdded   time.Time `bson:"date_added" json:"date_added"`
	OriginalUrl string    `bson:"original_url" json:"original_url"`
	Version     int       `bson:"version" json:"version"`
}

func buildUrl(rowId int, from time.Time, to time.Time) string {
	fromYear := from.Year()
	fromMonth := from.Month()
	fromDay := from.Day()

	toYear := to.Year()
	toMonth := to.Month()
	toDay := to.Day()

	return fmt.Sprintf(urlTemplate, rowId, fromYear, fromMonth, fromDay, toYear, toMonth, toDay, fromYear, fromMonth, fromDay, toYear, toMonth, toDay)
}

func FetchPublicationsPage(client *http.Client, row int, from time.Time, to time.Time) (result FetchedPublicationPage, err error) {
	url := buildUrl(row, from, to)

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

	if !strings.Contains(raw, "Einde van de lijst") {
		result.Raw = raw
		result.OriginalUrl = url
		result.DateAdded = time.Now()
		result.Version = 1
	}

	return result, err
}
