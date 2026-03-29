package repository

import "PolyakovEvg/go-musthave-shortener-tpl/internal/model"

type Repository interface {
	Save(string, string) (string, error)
	SaveBatch(userID string, batch []model.BatchRequest) ([]model.BatchResponse, error)
	Ping() error
	Get(string) (string, bool)
	GetByUser(userID string) ([]model.URL, error)
}
