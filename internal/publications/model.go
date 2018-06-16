package publications

import "gopkg.in/mgo.v2/bson"

type Publication struct {
	ID              bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	CompanyName     string        `bson:"company_name" json:"company_name"`
	LegalForm       string        `bson:"legal_form" json:"legal_form"`
	Address         string        `bson:"address" json:"address"`
	FileLocation    string        `bson:"file_location" json:"file_location"`
	DossierNumber   string        `bson:"dossier_number" json:"dossier_number"`
	DatePublication string        `bson:"date_publication" json:"date_publication"`
	Type            string        `bson:"type" json:"type"`
}
