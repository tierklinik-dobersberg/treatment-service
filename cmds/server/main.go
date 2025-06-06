package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/bufbuild/protovalidate-go"
	"github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1/treatmentv1connect"
	"github.com/tierklinik-dobersberg/apis/pkg/auth"
	"github.com/tierklinik-dobersberg/apis/pkg/cors"
	"github.com/tierklinik-dobersberg/apis/pkg/log"
	"github.com/tierklinik-dobersberg/apis/pkg/server"
	"github.com/tierklinik-dobersberg/apis/pkg/validator"
	"github.com/tierklinik-dobersberg/treatment-service/internal/config"
	"github.com/tierklinik-dobersberg/treatment-service/internal/service"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func main() {
	ctx := context.Background()

	var cfgFilePath string
	if len(os.Args) > 1 {
		cfgFilePath = os.Args[1]
	}

	cfg, err := config.LoadConfig(ctx, cfgFilePath)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration loaded successfully")

	providers, err := config.NewProviders(ctx, *cfg)
	if err != nil {
		slog.Error("failed to prepare providers", "error", err)
		os.Exit(1)
	}

	slog.Info("application providers prepared successfully")

	protoValidator, err := protovalidate.New()
	if err != nil {
		slog.Error("failed to prepare protovalidate", "error", err)
		os.Exit(1)
	}

	authInterceptor := auth.NewAuthAnnotationInterceptor(
		protoregistry.GlobalFiles,
		auth.NewIDMRoleResolver(providers.Roles),
		auth.RemoteHeaderExtractor)

	interceptors := connect.WithInterceptors(
		log.NewLoggingInterceptor(),
		authInterceptor,
		validator.NewInterceptor(protoValidator),
	)

	corsConfig := cors.Config{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowCredentials: true,
	}

	// Prepare our servemux and add handlers.
	serveMux := http.NewServeMux()

	// create a new CallService and add it to the mux.
	svc := service.New(providers)

	path, handler := treatmentv1connect.NewSpeciesServiceHandler(svc, interceptors)
	serveMux.Handle(path, handler)

	path, handler = treatmentv1connect.NewTreatmentServiceHandler(svc, interceptors)
	serveMux.Handle(path, handler)

	// Create the server
	srv, err := server.CreateWithOptions(cfg.PublicListenAddress, cors.Wrap(corsConfig, serveMux))
	if err != nil {
		slog.Error("failed to configure server", "error", err)
	}

	slog.Info("HTTP/2 server (h2c) prepared successfully, startin to listen ...")

	if err := server.Serve(ctx, srv); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
