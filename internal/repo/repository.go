package repo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	species    *mongo.Collection
	treatments *mongo.Collection

	initialTimeRequirement    time.Duration
	additionalTimeRequirement time.Duration
}

func NewRepositoryWithClient(ctx context.Context, cli *mongo.Client, databaseName string, defaultInitialTimeRequirement, defaultAdditionalTimeRequirement time.Duration) (*Repository, error) {
	db := cli.Database(databaseName)

	r := &Repository{
		species:    db.Collection("species"),
		treatments: db.Collection("treatments"),

		initialTimeRequirement:    defaultInitialTimeRequirement,
		additionalTimeRequirement: defaultAdditionalTimeRequirement,
	}

	if err := r.setup(ctx); err != nil {
		return nil, err
	}

	return r, nil
}

func NewRepository(ctx context.Context, databaseURL, databaseName string, defaultInitialTimeRequirement, defaultAdditionalTimeRequirement time.Duration) (*Repository, error) {
	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(databaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := cli.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return NewRepositoryWithClient(ctx, cli, databaseName, defaultInitialTimeRequirement, defaultAdditionalTimeRequirement)
}

func (r *Repository) setup(ctx context.Context) error {
	if _, err := r.species.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "matchWords", Value: 1},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	if _, err := r.treatments.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "species", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "matchEventText", Value: 1},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

func (r *Repository) withTransaction(ctx context.Context, fn func(mongo.SessionContext) (any, error)) (any, error) {
	session, err := r.treatments.Database().Client().StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer session.EndSession(ctx)

	return session.WithTransaction(ctx, fn)
}
