package config

import (
	"context"
	"fmt"

	"github.com/tierklinik-dobersberg/apis/pkg/discovery/wellknown"
	"github.com/tierklinik-dobersberg/apis/pkg/service"
	"github.com/tierklinik-dobersberg/treatment-service/internal/repo"
	"go.mongodb.org/mongo-driver/mongo"
)

type Providers struct {
	Clients *wellknown.Clients

	Repository *repo.Repository

	Config Config
}

type Instance = service.Instance[Config, *mongo.Database]

func NewProviders(ctx context.Context, i *Instance) (*Providers, error) {
	repo, err := repo.NewRepositoryWithClient(ctx, i.Database, i.Config.DefaultInitialTimeRequirement, i.Config.DefaultAdditionalTimeRequirement)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	p := &Providers{
		Clients:    &i.Clients,
		Repository: repo,
		Config:     i.Config,
	}

	return p, nil
}
