package companies

import "time"

type Company struct {
	General         General
	LegalFunctions  []Function
	Characteristics []Characteristic
	Licenses        []License
	Financial       Financial
	Related         []interface{}
	Activities      []interface{}
}

type General struct {
	CompanyNumber      string
	Active             string
	LegalStatus        LegalStatus
	StartDate          time.Time
	Name               LegalName
	Abbreviation       Abbreviation
	Address            Address
	TelephoneNumber    string
	FaxNumber          string
	Email              string
	Website            string
	Type               string
	LegalForm          LegalForm
	EstablishmentCount int
}

type LegalStatus struct {
	Situation  string
	Supplement string
}

type LegalName struct {
	Name       string
	Supplement string
}

type LegalForm struct {
	Form       string
	Supplement string
}

type Abbreviation struct {
	Name       string
	Supplement string
}

type Address struct {
	Street       string
	StreetNumber int
	PostalCode   string
	Place        string
	Country      string
}

type Function struct {
	Description string
	Name        string
	StartDate   time.Time
}

type Characteristic struct {
}

type License struct {
}

type Financial struct {
	AuthorizedCapital Money
	AnnualAssembly    string
	FinancialYearEnd  string
}

type Money struct {
	Amount   float32
	Currency string
}
