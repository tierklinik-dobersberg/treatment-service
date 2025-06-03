package service

import "github.com/tierklinik-dobersberg/treatment-service/internal/config"

type Service struct {
	*config.Providers
}

func New(p *config.Providers) *Service {
	return &Service{
		Providers: p,
	}
}
