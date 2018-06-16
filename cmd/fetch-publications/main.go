package main

import (
	"fmt"
	"github.com/demeyerthom/belgian-companies/internal/publications"
	"github.com/vjeantet/jodaTime"
	"io/ioutil"
	"os"
	"time"
)

var row = 1
var from = time.Now().AddDate(0, 0, -1)
var to = time.Now().AddDate(0, 0, -1)
var FileLocation = "/tmp/belgian-companies/lists/pages"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	isSuccess := true

	for isSuccess {
		result, err := publications.FetchPublicationsPage(row, from, to)

		if err != nil {
			isSuccess = false
			fmt.Println("Finished fetching")
			os.Exit(0)
		}

		err = ioutil.WriteFile(
			fmt.Sprintf("%s/%d-%s.html", FileLocation, row, jodaTime.Format("YYYY.MM.dd", from)),
			result,
			0644,
		)
		check(err)

		row = row + 30
	}
}
