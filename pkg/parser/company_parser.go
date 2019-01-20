package parser

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	"strings"
)

var (
	noDataIncluded = "No data included in CBE."

	placeAndPostalCodeRegex = regexp.MustCompile(`^(?P<postalcode>[0-9]{4}).(?P<place>.*)$`)
)

type CompanyParser struct {
	*Parser
	sanitizer *bluemonday.Policy
}

func NewCompanyParser() *CompanyParser {
	sanitizer := bluemonday.UGCPolicy()
	return &CompanyParser{sanitizer: sanitizer}
}

func (p *CompanyParser) ParseCompanyPages(pages *model.CompanyPages) (*model.Company, error) {
	company := model.NewCompany()

	doc, err := htmlquery.Parse(strings.NewReader(pages.Company.Raw))
	if err != nil {
		return nil, err
	}

	company.DossierNumber, err = p.parseCompanyEnterpriseNumber(doc)
	if err != nil {
		return nil, err
	}

	company.VATNumber, err = p.parseCompanyEnterpriseNumber(doc)
	if err != nil {
		return nil, err
	}

	company.Status, err = p.parseStatus(doc)
	if err != nil {
		return nil, err
	}

	company.LegalSituation, err = p.parseLegalSituation(doc)
	if err != nil {
		return nil, err
	}

	company.StartDate, err = p.parseStartDate(doc)
	if err != nil {
		return nil, err
	}

	company.LegalName, err = p.parseLegalName(doc)
	if err != nil {
		return nil, err
	}

	company.PhoneNumber, err = p.parsePhoneNumber(doc)
	if err != nil {
		return nil, err
	}

	company.FaxNumber, err = p.parseFaxNumber(doc)
	if err != nil {
		return nil, err
	}

	company.WebAddress, err = p.parseWebAddress(doc)
	if err != nil {
		return nil, err
	}

	company.EmailAddress, err = p.parseEmailAddress(doc)
	if err != nil {
		return nil, err
	}

	company.LegalType, err = p.parseLegalType(doc)
	if err != nil {
		return nil, err
	}

	company.LegalForm, err = p.parseLegalForm(doc)
	if err != nil {
		return nil, err
	}

	company.HeadOfficeAddress, err = p.parseHeadOfficeAddress(doc)
	if err != nil {
		return nil, err
	}

	company.LegalFunctions, err = p.parseLegalFunctions(doc)

	return company, nil
}

func (p *CompanyParser) parseCompanyEnterpriseNumber(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'QL') and text()[contains(.,'Enterprise number:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find enterprise number")
	}

	return strings.Replace(strings.TrimSpace(htmlquery.InnerText(nodes[0])), ".", "", 10), nil
}

func (p *CompanyParser) parseStatus(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'RL') and text()[contains(.,'Status:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find status")
	}

	return strings.ToLower(strings.TrimSpace(htmlquery.InnerText(nodes[0]))), nil
}

func (p *CompanyParser) parseLegalSituation(node *html.Node) (item *model.DatedItem, err error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'QL') and text()[contains(.,'Legal situation:')]]/../td[2]")
	if len(nodes) != 1 {
		return nil, errors.New("could not find legal situation")
	}

	item = model.NewDatedItem()

	statusNodes := htmlquery.Find(nodes[0], "//strong/span")
	if len(statusNodes) != 1 {
		return nil, errors.New("could not find legal situation name")
	}
	item.Text = strings.ToLower(strings.TrimSpace(htmlquery.InnerText(statusNodes[0])))

	dateNodes := htmlquery.Find(nodes[0], "//span[@class='upd']")
	if len(dateNodes) != 1 {
		return nil, errors.New("could not find legal situation date")
	}

	item.DateFrom = strings.ToLower(strings.TrimSpace(htmlquery.InnerText(dateNodes[0])))

	return item, err
}

func (p *CompanyParser) parseStartDate(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'RL') and text()[contains(.,'Start date:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find start date")
	}

	return strings.ToLower(strings.TrimSpace(htmlquery.InnerText(nodes[0]))), nil
}

func (p *CompanyParser) parseLegalName(node *html.Node) (item *model.DatedItem, err error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'QL') and text()[contains(.,'Legal name:')]]/../td[2]")
	if len(nodes) != 1 {
		return nil, errors.New("could not find legal name")
	}

	item = model.NewDatedItem()

	nameNodes := htmlquery.Find(nodes[0], "/text()")
	if len(nameNodes) < 1 && len(nameNodes) > 2 {
		return nil, errors.New("could not find legal name text")
	}
	item.Text = strings.TrimSpace(htmlquery.InnerText(nameNodes[0]))

	dateNodes := htmlquery.Find(nodes[0], "//span[@class='upd']")
	if len(dateNodes) != 1 {
		return nil, errors.New("could not find legal name date")
	}

	item.DateFrom = strings.TrimSpace(htmlquery.InnerText(dateNodes[0]))

	return item, err
}

func (p *CompanyParser) parsePhoneNumber(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'QL') and text()[contains(.,'Phone number:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find phone number")
	}

	if strings.TrimSpace(htmlquery.InnerText(nodes[0])) == noDataIncluded {
		return "", nil
	}

	return strings.TrimSpace(htmlquery.InnerText(nodes[0])), nil
}

func (p *CompanyParser) parseFaxNumber(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'RL') and text()[contains(.,'Fax:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find fax number")
	}

	if strings.TrimSpace(htmlquery.InnerText(nodes[0])) == noDataIncluded {
		return "", nil
	}

	return strings.TrimSpace(htmlquery.InnerText(nodes[0])), nil
}

func (p *CompanyParser) parseWebAddress(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'RL') and text()[contains(.,'Web Address:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find web address")
	}

	if strings.TrimSpace(htmlquery.InnerText(nodes[0])) == noDataIncluded {
		return "", nil
	}

	return strings.TrimSpace(htmlquery.InnerText(nodes[0])), nil
}

func (p *CompanyParser) parseEmailAddress(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'QL') and text()[contains(.,'Email address:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find email address")
	}

	if strings.TrimSpace(htmlquery.InnerText(nodes[0])) == noDataIncluded {
		return "", nil
	}

	return strings.TrimSpace(htmlquery.InnerText(nodes[0])), nil
}

func (p *CompanyParser) parseLegalType(node *html.Node) (string, error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'QL') and text()[contains(.,'Entity type:')]]/../td[2]")
	if len(nodes) != 1 {
		return "", errors.New("could not find legal type")
	}

	return strings.TrimSpace(htmlquery.InnerText(nodes[0])), nil
}

func (p *CompanyParser) parseLegalForm(node *html.Node) (item *model.DatedItem, err error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'RL') and text()[contains(.,'Legal form:')]]/../td[2]")
	if len(nodes) != 1 {
		return nil, errors.New("could not find legal form")
	}

	item = model.NewDatedItem()

	nameNodes := htmlquery.Find(nodes[0], "/text()")
	if len(nameNodes) < 1 && len(nameNodes) > 2 {
		return nil, errors.New("could not find legal form text")
	}
	item.Text = strings.TrimSpace(htmlquery.InnerText(nameNodes[0]))

	dateNodes := htmlquery.Find(nodes[0], "//span[@class='upd']")
	if len(dateNodes) != 1 {
		return nil, errors.New("could not find legal form date")
	}

	item.DateFrom = strings.TrimSpace(htmlquery.InnerText(dateNodes[0]))

	return item, err
}

func (p *CompanyParser) parseHeadOfficeAddress(node *html.Node) (item *model.DatedAddress, err error) {
	nodes := htmlquery.Find(node, "//div[@id='table']/table//td[contains(@class, 'RL') and text()[contains(.,'Head office')]]/../td[2]")
	if len(nodes) != 1 {
		return nil, errors.New("could not find head office address")
	}

	item = model.NewDatedAddress()

	dateNodes := htmlquery.Find(nodes[0], "//span[@class='upd']")
	if len(dateNodes) != 1 {
		return nil, errors.New("could not find head office since date")
	}

	item.DateFrom = strings.TrimSpace(htmlquery.InnerText(dateNodes[0]))

	address := model.NewAddress()

	addressNodes := htmlquery.Find(nodes[0], "/text()")
	if len(addressNodes) < 1 && len(addressNodes) > 2 {
		return nil, errors.New("could not find address text")
	}
	address.Street = strings.TrimSpace(htmlquery.InnerText(addressNodes[0]))

	placeAndPostalCodeString := strings.TrimSpace(htmlquery.InnerText(addressNodes[1]))
	match := placeAndPostalCodeRegex.FindStringSubmatch(placeAndPostalCodeString)
	subMatchMap := make(map[string]string)
	for i, name := range placeAndPostalCodeRegex.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}
	address.Place = subMatchMap["place"]
	postalCode, _ := strconv.Atoi(subMatchMap["postalcode"])
	address.PostalCode = int32(postalCode)

	item.Address = address

	return item, err
}

func (p *CompanyParser) parseLegalFunctions(node *html.Node) ([]*model.LegalFunction, error) {
	nodes := htmlquery.Find(
		node,
		"//tr[td[contains(@class, 'I') and text()[contains(.,'Legal functions')]]]"+
			"/following-sibling::tr[td[not(contains(@class, 'I'))]]",
	)
	//tr/

	for _, node := range nodes {
		text := htmlquery.OutputHTML(node, true)
		fmt.Print(text)
	}

	return nil, nil
}
