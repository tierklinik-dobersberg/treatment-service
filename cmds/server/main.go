package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1/treatmentv1connect"
	"github.com/tierklinik-dobersberg/apis/pkg/discovery/wellknown"
	base "github.com/tierklinik-dobersberg/apis/pkg/service"
	"github.com/tierklinik-dobersberg/treatment-service/internal/config"
	"github.com/tierklinik-dobersberg/treatment-service/internal/service"
)

var serverContextKey = struct{ S string }{S: "serverContextKey"}

func main() {
	ctx := context.Background()

	instance, err := base.Configure(
		wellknown.TreatmentV1ServiceScope,
		config.Config{},
	)
	if err != nil {
		slog.Error("failed to configure service instance: %w", err)
	}

	providers, err := config.NewProviders(ctx, instance)
	if err != nil {
		slog.Error("failed to prepare providers", "error", err)
		os.Exit(1)
	}

	slog.Info("application providers prepared successfully")

	// create a new CallService and add it to the mux.
	svc := service.New(providers)

	path, handler := treatmentv1connect.NewSpeciesServiceHandler(svc, connect.WithOptions(instance.ConnectOptions()...))
	instance.Mux.Shared.Handle(path, handler)

	path, handler = treatmentv1connect.NewTreatmentServiceHandler(svc, connect.WithOptions(instance.ConnectOptions()...))
	instance.Mux.Shared.Handle(path, handler)

	slog.Info("HTTP/2 server (h2c) prepared successfully, starting to listen ...")

	if err := instance.Run(); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
