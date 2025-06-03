package config

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tierklinik-dobersberg/apis/gen/go/tkd/idm/v1/idmv1connect"
	"github.com/tierklinik-dobersberg/apis/pkg/h2utils"
	"github.com/tierklinik-dobersberg/treatment-service/internal/repo"
)

type Providers struct {
	Users idmv1connect.UserServiceClient
	Roles idmv1connect.RoleServiceClient

	Repository *repo.Repository

	Config Config
}

func NewProviders(ctx context.Context, cfg Config) (*Providers, error) {
	httpClient := h2utils.WithDiscovery(nil, http.DefaultClient)

	repo, err := repo.NewRepository(ctx, cfg.MongoURL, cfg.DatabaseName, cfg.DefaultInitialTimeRequirement, cfg.DefaultAdditionalTimeRequirement)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	p := &Providers{
		Users:      idmv1connect.NewUserServiceClient(httpClient, cfg.IdmURL),
		Roles:      idmv1connect.NewRoleServiceClient(httpClient, cfg.IdmURL),
		Repository: repo,
		Config:     cfg,
	}

	return p, nil
}
