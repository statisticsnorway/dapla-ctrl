package atlantis

import (
	"context"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

func Create(ctx context.Context, teamSlug slug.Slug) error {
	return db(ctx).Create(ctx, teamSlug)
}
