package seed

import (
	"context"
	"log"
)

// backfillMissingCRMProfiles ensures every customer profile has customerType and
// purchasedCategories in import_profile (generated crm_* columns follow automatically).
func (r *Runner) backfillMissingCRMProfiles(ctx context.Context) error {
	log.Println("seed: backfilling missing CRM profile fields...")
	const q = `
UPDATE customer_profiles cp
SET import_profile = COALESCE(cp.import_profile, '{}'::jsonb)
  || jsonb_build_object(
    'customerType', COALESCE(NULLIF(cp.import_profile->>'customerType', ''), 'public')
  )
  || CASE
    WHEN jsonb_typeof(COALESCE(cp.import_profile->'purchasedCategories', '[]'::jsonb)) = 'array'
     AND jsonb_array_length(COALESCE(cp.import_profile->'purchasedCategories', '[]'::jsonb)) > 0
    THEN '{}'::jsonb
    ELSE jsonb_build_object('purchasedCategories', '["gold_and_stones"]'::jsonb)
  END,
  import_mode = COALESCE(cp.import_mode, 'history_included'::customer_import_mode),
  updated_at = NOW()
WHERE cp.import_profile IS NULL
   OR NULLIF(cp.import_profile->>'customerType', '') IS NULL
   OR jsonb_array_length(COALESCE(cp.import_profile->'purchasedCategories', '[]'::jsonb)) = 0`
	tag, err := r.pool.Exec(ctx, q)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		log.Printf("seed:   backfilled CRM fields on %d profiles", tag.RowsAffected())
	}
	return nil
}
