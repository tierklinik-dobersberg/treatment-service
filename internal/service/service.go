package service

import (
	"github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1/treatmentv1connect"
	"github.com/tierklinik-dobersberg/treatment-service/internal/config"
)

type Service struct {
	*config.Providers

	treatmentv1connect.UnimplementedSpeciesServiceHandler
	treatmentv1connect.UnimplementedTreatmentServiceHandler
}

func New(p *config.Providers) *Service {
	return &Service{
		Providers: p,
	}
}
