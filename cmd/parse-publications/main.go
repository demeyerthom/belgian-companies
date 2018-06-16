package main

import (
	"fmt"
	"github.com/demeyerthom/belgian-companies/internal/publications"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"os"
)

var FileLocation = "/tmp/belgian-companies/lists/pages"
var ParsedLocation = "/tmp/belgian-companies/lists/parsed"
var DocumentsLocation = "/tmp/belgian-companies/files/publications"
var WithDocuments = false

func main() {
	files, _ := ioutil.ReadDir(FileLocation)

	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	publicationsCollection := session.DB("belgian-companies").C("publications")

	for _, f := range files {
		newPublications, err := publications.ParsePublicationPage(FileLocation+"/"+f.Name(), DocumentsLocation, WithDocuments)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range newPublications {
			fmt.Printf("%+v\n", p)
			err = publicationsCollection.Insert(p)
			if err != nil {
				panic(err)
			}
		}

		os.Rename(FileLocation+"/"+f.Name(), ParsedLocation+"/"+f.Name())
	}
}
