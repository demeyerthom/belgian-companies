package storage

import (
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/demeyerthom/belgian-companies/pkg/util"
	"time"
)

type Record struct {
	DossierNumber string
	date          time.Time
}

type Storage struct {
	Adapter Adapter
}

func NewStorage(adapter Adapter) *Storage {
	return &Storage{Adapter: adapter}
}

func (s *Storage) ShouldProcess(publication *model.Publication) (bool, error) {
	record, err := s.Adapter.GetRecord(publication)
	util.Check(err)

	if record == nil {
		return false, err
	}

	if record.date > publication.DatePublication {
		return false, err
	}

	return true, err
}

func (s *Storage) ShouldNotProcess(publication *model.Publication) (bool, error) {
	shouldProcess, err := s.ShouldProcess(publication)
	return !shouldProcess, err
}

func (s *Storage) Update(publication *model.Publication) error {
	return nil
}

func (s *Storage) Close() {

}
