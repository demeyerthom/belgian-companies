package fetcher

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	assetsDir = "../../test/data/company-pages/"
)

type mockHttpClient struct {
	mock.Mock
}

type call struct {
	url          string
	responseFile string
	responseCode int
	err          error
}

func (o *mockHttpClient) Get(url string) (resp *http.Response, err error) {
	args := o.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

func makeMockCompanyFetcher(calls []call) *CompanyFetcher {
	httpClient := new(mockHttpClient)

	for _, call := range calls {
		page, err := ioutil.ReadFile(assetsDir + call.responseFile)
		if err != nil {
			panic(err)
		}

		reader := bytes.NewBuffer(page)

		response := http.Response{
			StatusCode: call.responseCode,
			Body:       ioutil.NopCloser(reader),
		}

		httpClient.On("Get", call.url).Return(&response, call.err)
	}

	return NewCompanyFetcher(httpClient, 1)
}

func TestCompanyFetcher_FetchCompanyPagesWithSingleEstablishment(t *testing.T) {
	assert := assert.New(t)

	companyId := "403170701"
	establishmentId := "2269751597"

	var calls []call

	companyCall := call{
		url:          fmt.Sprintf(companyMainPageTemplate, companyId),
		responseFile: "684446351-main.html",
		responseCode: 200,
		err:          nil,
	}

	listCall := call{
		url:          fmt.Sprintf(establishmentPageListTemplate, companyId, 1),
		responseFile: "684446351-establishment-list.html",
		responseCode: 200,
		err:          nil,
	}

	establishmentCall := call{
		url:          rootPageTemplate + "toonvestigingps.html?vestigingsnummer=" + establishmentId + "&lang=en",
		responseFile: "684446351-establishment-2269751597.html",
		responseCode: 200,
		err:          nil,
	}
	calls = append(calls, companyCall, listCall, establishmentCall)

	fetcher := makeMockCompanyFetcher(calls)
	pages, err := fetcher.FetchCompanyPages(companyId)

	assert.Nil(err)
	assert.NotNil(pages.Company)
	assert.Len(pages.Establishments, 1)
}

func TestCompanyFetcher_FetchCompanyPagesNoEstablishments(t *testing.T) {
	assert := assert.New(t)

	companyId := "679873394"

	var calls []call

	companyUrl := fmt.Sprintf(companyMainPageTemplate, companyId)
	companyCall := call{
		url:          companyUrl,
		responseFile: "679873394-main-no-establishments.html",
		responseCode: 200,
		err:          nil,
	}
	calls = append(calls, companyCall)

	fetcher := makeMockCompanyFetcher(calls)
	pages, err := fetcher.FetchCompanyPages(companyId)

	assert.Nil(err)
	assert.NotNil(pages.Company)
	assert.Len(pages.Establishments, 0)
}

//func TestCompanyFetcher_FetchCompanyPagesManyEstablishments(t *testing.T) {
//	assert := assert.New(t)
//
//	listPageCount := 24
//	establishmentIds := []string{
//		"2104586333",
//		"2104586432",
//		"2104587125",
//		"2104587521",
//		"2104587818",
//		"2104588214",
//		"2104588412",
//		"2104588511",
//		"2104588610",
//		"2104588907",
//		"2104589105",
//		"2104589895",
//		"2104590093",
//		"2104590291",
//		"2104591776",
//		"2104592469",
//		"2104592766",
//		"2104592865",
//		"2104592964",
//		"2104594350",
//	}
//	companyId := "403199702"
//
//	var calls []call
//
//	companyCall := call{
//		url:          fmt.Sprintf(companyMainPageTemplate, companyId),
//		responseFile: "403199702-main.html",
//		responseCode: 200,
//		err:          nil,
//	}
//	calls = append(calls, companyCall)
//
//	var listCalls []call
//	for i := 0; i < listPageCount; i++ {
//		listCall := call{
//			url:          fmt.Sprintf(establishmentPageListTemplate, companyId, 1),
//			responseFile: "403199702-establishment-list-page-1.html",
//			responseCode: 200,
//			err:          nil,
//		}
//		var establishmentCalls []call
//		for i := 0; i < 20; i++ {
//			establishmentCall := call{
//				url:          fmt.Sprintf(rootPageTemplate+"toonvestigingps.html?vestigingsnummer=%s&lang=en", establishmentIds[i]),
//				responseFile: "684446351-establishment-2269751597.html",
//				responseCode: 200,
//				err:          nil,
//			}
//			establishmentCalls = append(establishmentCalls, establishmentCall)
//		}
//		listCalls = append(listCalls, listCall)
//		calls = append(calls, establishmentCalls...)
//	}
//	calls = append(calls, listCalls...)
//
//	fetcher := makeMockCompanyFetcher(calls)
//	pages, err := fetcher.FetchCompanyPages(companyId)
//
//	assert.Nil(err)
//	assert.Len(pages.Pages, 1)
//}
