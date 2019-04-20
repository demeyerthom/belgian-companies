package storage

import "github.com/demeyerthom/belgian-companies/pkg/model"

type Adapter interface {
	Close() error
	GetRecord(publication *model.Publication) (record *Record, err error)
	//UpdateRecord(publication model.Publication) error
}
