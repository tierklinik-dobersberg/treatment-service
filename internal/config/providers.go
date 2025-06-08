package config

import (
	"context"
	"fmt"

	"github.com/tierklinik-dobersberg/apis/pkg/discovery"
	"github.com/tierklinik-dobersberg/apis/pkg/discovery/wellknown"
	"github.com/tierklinik-dobersberg/treatment-service/internal/repo"
)

type Providers struct {
	Clients *wellknown.Clients

	Repository *repo.Repository

	Config Config
}

func NewProviders(ctx context.Context, cfg Config, catalog discovery.Discoverer) (*Providers, error) {
	repo, err := repo.NewRepository(ctx, cfg.MongoURL, cfg.DatabaseName, cfg.DefaultInitialTimeRequirement, cfg.DefaultAdditionalTimeRequirement)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	clients := wellknown.ConfigureClients(wellknown.ConfigureClientOptions{
		Catalog: catalog,
	})

	p := &Providers{
		Clients:    &clients,
		Repository: repo,
		Config:     cfg,
	}

	return p, nil
}
