package mocks

import (
	"leadsextractor/models"
	"leadsextractor/store"
)

type MockUTMStorer struct {
	utms []*models.UtmDefinition
}

func (s *MockUTMStorer) Mock() {
	utm := models.UtmDefinition{
		Code:     "CODE1",
		Source:   models.NullString{String: "aa", Valid: true},
		Medium:   models.NullString{String: "aa", Valid: true},
		Campaign: models.NullString{String: "aa", Valid: true},
		Ad:       models.NullString{String: "aa", Valid: true},
		Channel:  models.NullString{String: "ivr", Valid: true},
	}
	s.Insert(&utm)
}

func (s *MockUTMStorer) Utms() []*models.UtmDefinition {
	return s.utms
}

// GetAll(*[]models.UtmDefinition) error
func (s *MockUTMStorer) GetAll() ([]*models.UtmDefinition, error) {
	return s.utms, nil
}

// GetOne(int) (*models.UtmDefinition, error)
func (s *MockUTMStorer) GetOne(id int) (*models.UtmDefinition, error) {
	for _, utm := range s.utms {
		if utm.Id == id {
			return utm, nil
		}
	}
	return nil, store.NewErr("not found", store.StoreNotFoundErr)
}

// GetOneByCode(string) (*models.UtmDefinition, error)
func (s *MockUTMStorer) GetOneByCode(code string) (*models.UtmDefinition, error) {
	for _, utm := range s.utms {
		if utm.Code == code {
			return utm, nil
		}
	}
	return nil, store.NewErr("not found", store.StoreNotFoundErr)
}

func (s *MockUTMStorer) Insert(utm *models.UtmDefinition) (int64, error) {
	if u, _ := s.GetOneByCode(utm.Code); u != nil {
		return 0, store.NewErr("already exists", store.StoreDuplicatedErr)
	}

	s.utms = append(s.utms, utm)
	id := len(s.utms)
	s.utms[id-1].Id = id
	return int64(id), nil
}

// Update(*models.UtmDefinition) error
func (s *MockUTMStorer) Update(uUTM *models.UtmDefinition) error {
	if u, _ := s.GetOneByCode(uUTM.Code); u != nil {
		return store.NewErr("already exists", store.StoreDuplicatedErr)
	}

	s.utms[uUTM.Id-1] = uUTM
	return nil
}

func (s *MockUTMStorer) Delete(id int) error {
	newList := make([]*models.UtmDefinition, 0)
	for _, utm := range s.utms {
		if utm.Id == id {
			continue
		}
		newList = append(newList, utm)
	}
	if len(s.utms) == len(newList) {
		return store.NewErr("not found", store.StoreNotFoundErr)
	}

	s.utms = newList
	return nil
}
