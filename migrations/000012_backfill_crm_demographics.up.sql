-- Backfill missing CRM demographics on existing customer import profiles.
-- Generated crm_age_range / crm_gender columns update automatically.

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
  END
  || CASE
    WHEN NULLIF(cp.import_profile->>'customerAgeRange', '') IS NULL
    THEN jsonb_build_object(
      'customerAgeRange',
      (ARRAY['1-7', '7-14', '14-21', '21-40', '40+'])[1 + (abs(hashtext(cp.user_id::text)) % 5)]
    )
    ELSE '{}'::jsonb
  END
  || CASE
    WHEN NULLIF(cp.import_profile->>'gender', '') IS NULL
     AND abs(hashtext(cp.user_id::text || ':gender')) % 7 <> 0
    THEN jsonb_build_object(
      'gender',
      (ARRAY['male', 'female', 'other'])[1 + (abs(hashtext(cp.user_id::text || ':gender')) % 3)]
    )
    ELSE '{}'::jsonb
  END,
  import_mode = COALESCE(cp.import_mode, 'history_included'::customer_import_mode),
  updated_at = NOW()
WHERE cp.import_profile IS NULL
   OR NULLIF(cp.import_profile->>'customerType', '') IS NULL
   OR jsonb_array_length(COALESCE(cp.import_profile->'purchasedCategories', '[]'::jsonb)) = 0
   OR NULLIF(cp.import_profile->>'customerAgeRange', '') IS NULL
   OR (
     NULLIF(cp.import_profile->>'gender', '') IS NULL
     AND abs(hashtext(cp.user_id::text || ':gender')) % 7 <> 0
   );
