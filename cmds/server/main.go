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
	"github.com/tierklinik-dobersberg/apis/pkg/discovery"
	"github.com/tierklinik-dobersberg/apis/pkg/discovery/consuldiscover"
	"github.com/tierklinik-dobersberg/apis/pkg/discovery/wellknown"
	"github.com/tierklinik-dobersberg/apis/pkg/log"
	"github.com/tierklinik-dobersberg/apis/pkg/server"
	"github.com/tierklinik-dobersberg/apis/pkg/validator"
	"github.com/tierklinik-dobersberg/treatment-service/internal/config"
	"github.com/tierklinik-dobersberg/treatment-service/internal/service"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var serverContextKey = struct{ S string }{S: "serverContextKey"}

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
		func(ctx context.Context, req connect.AnyRequest) (auth.RemoteUser, error) {
			serverKey, _ := ctx.Value(serverContextKey).(string)

			if serverKey == "admin" {
				return auth.RemoteUser{
					ID:          "service-account",
					DisplayName: req.Peer().Addr,
					RoleIDs:     []string{"idm_superuser"}, // FIXME(ppacher): use a dedicated manager role for this
					Admin:       true,
				}, nil
			}

			return auth.RemoteHeaderExtractor(ctx, req)
		},
	)

	interceptors := []connect.Interceptor{
		log.NewLoggingInterceptor(),
		validator.NewInterceptor(protoValidator),
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	if os.Getenv("DEBUG") == "" {
		interceptors = append(interceptors, authInterceptor)
	}

	corsConfig := cors.Config{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowCredentials: true,
	}

	wrapWithKey := func(key string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), serverContextKey, key))

			next.ServeHTTP(w, r)
		})
	}

	// Register at service catalog
	catalog, err := consuldiscover.NewFromEnv()
	if err != nil {
		slog.Error("failed to get service catalog client", "error", err)
		os.Exit(1)
	}

	if err := discovery.Register(ctx, catalog, &discovery.ServiceInstance{
		Name:    wellknown.TreatmentV1ServiceScope,
		Address: cfg.AdminListenAddress,
	}); err != nil {
		slog.Error("failed to register treatment-service at service catalog", "error", err)
	}

	// Prepare our servemux and add handlers.
	serveMux := http.NewServeMux()

	// create a new CallService and add it to the mux.
	svc := service.New(providers)

	path, handler := treatmentv1connect.NewSpeciesServiceHandler(svc, connect.WithInterceptors(interceptors...))
	serveMux.Handle(path, handler)

	path, handler = treatmentv1connect.NewTreatmentServiceHandler(svc, connect.WithInterceptors(interceptors...))
	serveMux.Handle(path, handler)

	// Create the servers
	srv, err := server.CreateWithOptions(cfg.PublicListenAddress, wrapWithKey("public", serveMux), server.WithCORS(corsConfig))
	if err != nil {
		slog.Error("failed to configure server", "error", err)
	}
	adminSrv, err := server.CreateWithOptions(cfg.AdminListenAddress, wrapWithKey("admin", serveMux), server.WithCORS(corsConfig))
	if err != nil {
		slog.Error("failed to configure server", "error", err)
	}

	slog.Info("HTTP/2 server (h2c) prepared successfully, starting to listen ...")

	if err := server.Serve(ctx, srv, adminSrv); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
