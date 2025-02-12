package handlers_test

import (
	"leadsextractor/models"
	"leadsextractor/store"
)


type mockCommStorer struct {
    comms []models.Communication
}

func (s *mockCommStorer) Insert(c *models.Communication, source *models.Source) error {
    return nil
}

func (s *mockCommStorer) GetAll(params *store.QueryParam) ([]models.Communication, error) {
    return s.comms, nil
}

func (s *mockCommStorer) GetDistinct(params *store.QueryParam) ([]models.Communication, error) {
    return nil, store.NewErr("does not exists", store.StoreNotFoundErr)
}

func (s *mockCommStorer) Count(params *store.QueryParam) (int, error) {
    return len(s.comms), nil
}

func (s *mockCommStorer) Exists(params *store.QueryParam) bool {
    return true
}
