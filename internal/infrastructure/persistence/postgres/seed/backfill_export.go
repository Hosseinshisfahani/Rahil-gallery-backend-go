package seed

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BackfillCRMProfiles fills missing CRM fields on customer import profiles.
func BackfillCRMProfiles(ctx context.Context, pool *pgxpool.Pool) error {
	r := &Runner{pool: pool}
	return r.backfillMissingCRMProfiles(ctx)
}
