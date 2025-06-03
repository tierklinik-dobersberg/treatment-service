package service

import (
	"context"

	"github.com/bufbuild/connect-go"
	treatmentv1 "github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1"
	"github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1/treatmentv1connect"
	"github.com/tierklinik-dobersberg/treatment-service/internal/config"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (svc *Service) CreateSpecies(ctx context.Context, req *connect.Request[treatmentv1.Species]) (*connect.Response[treatmentv1.Species], error) {
	res, err := svc.Repository.CreateSpecies(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (svc *Service) DeleteSpecies(ctx context.Context, req *connect.Request[treatmentv1.DeleteSpeciesRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := svc.Repository.DeleteSpecies(ctx, req.Msg.Name); err != nil {
		return nil, err
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (svc *Service) ListSpecies(ctx context.Context, req *connect.Request[treatmentv1.ListSpeciesRequest]) (*connect.Response[treatmentv1.ListSpeciesResponse], error) {
	res, err := svc.Repository.ListSpecies(ctx)
	if err != nil {
		return nil, err
	}

	response := &treatmentv1.ListSpeciesResponse{
		Species: res,
	}

	return connect.NewResponse(response), nil
}

func (svc *Service) UpdateSpecies(ctx context.Context, req *connect.Request[treatmentv1.UpdateSpeciesRequest]) (*connect.Response[treatmentv1.Species], error) {
	res, err := svc.Repository.UpdateSpecies(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
