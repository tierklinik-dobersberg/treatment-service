package repo

import "context"

type Repository struct {
}

func NewRepository(ctx context.Context, databaseURL string) (*Repository, error) {
	r := &Repository{}

	if err := r.setup(ctx); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) setup(ctx context.Context) error {
	return nil
}
