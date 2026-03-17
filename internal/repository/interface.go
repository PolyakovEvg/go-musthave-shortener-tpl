package repository

import "PolyakovEvg/go-musthave-shortener-tpl/internal/model"

type Repository interface {
	Save(string) (string, error)
	SaveBatch(batch []model.BatchRequest) ([]model.BatchResponse, error)
	Get(string) (string, bool)
}
