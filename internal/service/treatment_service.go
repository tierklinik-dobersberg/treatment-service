package service

import (
	"context"

	"github.com/bufbuild/connect-go"
	treatmentv1 "github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (svc *Service) CreateTreatment(ctx context.Context, req *connect.Request[treatmentv1.Treatment]) (*connect.Response[treatmentv1.Treatment], error) {
	res, err := svc.Repository.CreateTreatment(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (svc *Service) ListTreatments(ctx context.Context, req *connect.Request[treatmentv1.ListTreatmentsRequest]) (*connect.Response[treatmentv1.ListTreatmentsResponse], error) {
	species := []string{}
	if req.Msg.Species != "" {
		species = []string{req.Msg.Species}
	}

	res, err := svc.Repository.QuerySpecies(ctx, species, req.Msg.DisplayNameSearch)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&treatmentv1.ListTreatmentsResponse{
		Treatments: res,
	}), nil
}

func (svc *Service) UpdateTreatment(ctx context.Context, req *connect.Request[treatmentv1.UpdateTreatmentRequest]) (*connect.Response[treatmentv1.Treatment], error) {
	res, err := svc.Repository.UpdateTreatment(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (svc *Service) DeleteTreatment(ctx context.Context, req *connect.Request[treatmentv1.DeleteTreatmentRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := svc.Repository.DeleteTreatment(ctx, req.Msg.Name); err != nil {
		return nil, err
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (svc *Service) GetTreatment(ctx context.Context, req *connect.Request[treatmentv1.GetTreatmentRequest]) (*connect.Response[treatmentv1.Treatment], error) {
	res, err := svc.Repository.GetTreatment(ctx, req.Msg.Name)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
