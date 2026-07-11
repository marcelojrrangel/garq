package service

import "garq/internal/api"

type Service struct {
	api *api.API
}

func New(a *api.API) *Service {
	return &Service{api: a}
}
