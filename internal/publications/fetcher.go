package publications

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var urlTemplate = "http://www.ejustice.just.fgov.be/cgi_tsv/tsv_l_1.pl?lang=nl&row_id=%d&pdda=%d&pddm=%d&pddj=%d&pdfa=%d&pdfm=%d&pdfj=%d&fromtab=TSV_TMP&rech=1003&sql=pd+between+date%%27%d-%d-%d%%27+and+date%%27%d-%d-%d%%27+"

func FetchPublicationsPage(row int, from time.Time, to time.Time) (result []byte, err error) {
	url := buildUrl(row, from, to)

	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, errors.New("something went wrong")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	if bytes.Contains(body, []byte("Einde van de lijst")) {
		err = errors.New("end of list")
	}

	return body, err
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
