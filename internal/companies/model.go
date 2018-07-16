package companies

import "time"

type Company struct {
	CompanyNumber   string
	Active          bool
	LegalStatus     string
	FoundingDate    time.Time
	Name            string
	Address         Address
	TelephoneNumber string
	FaxNumber       string
	Email           string
	Website         string
	Type            string
	LegalForm       string
	Establishments  []Establishment
	Functions       []Function
}

type Address struct {
	Street       string
	StreetNumber int
	PostalCode   string
	Place        string
	Country      string
}

type Establishment struct {
}

type Function struct {
	Description string
}
