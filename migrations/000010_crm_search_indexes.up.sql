-- CRM advanced search: expression indexes on import_profile JSONB fields

CREATE INDEX IF NOT EXISTS idx_customer_profiles_import_age_range
    ON customer_profiles ((import_profile->>'customerAgeRange'))
    WHERE import_profile->>'customerAgeRange' IS NOT NULL
      AND import_profile->>'customerAgeRange' <> '';

CREATE INDEX IF NOT EXISTS idx_customer_profiles_import_gender
    ON customer_profiles ((import_profile->>'gender'))
    WHERE import_profile->>'gender' IS NOT NULL
      AND import_profile->>'gender' <> '';

CREATE INDEX IF NOT EXISTS idx_customer_profiles_import_customer_type
    ON customer_profiles ((import_profile->>'customerType'))
    WHERE import_profile->>'customerType' IS NOT NULL
      AND import_profile->>'customerType' <> '';

CREATE INDEX IF NOT EXISTS idx_customer_profiles_import_first_visit
    ON customer_profiles ((NULLIF(import_profile->>'firstVisitDate', '')::date))
    WHERE import_profile->>'firstVisitDate' IS NOT NULL
      AND import_profile->>'firstVisitDate' <> '';

CREATE INDEX IF NOT EXISTS idx_customer_profiles_import_birthday
    ON customer_profiles ((NULLIF(import_profile->>'birthday', '')::date))
    WHERE import_profile->>'birthday' IS NOT NULL
      AND import_profile->>'birthday' <> '';

CREATE INDEX IF NOT EXISTS idx_customer_profiles_import_marriage
    ON customer_profiles ((NULLIF(import_profile->>'marriageDate', '')::date))
    WHERE import_profile->>'marriageDate' IS NOT NULL
      AND import_profile->>'marriageDate' <> '';

CREATE INDEX IF NOT EXISTS idx_customer_profiles_import_categories
    ON customer_profiles USING gin ((import_profile->'purchasedCategories'))
    WHERE import_profile->'purchasedCategories' IS NOT NULL;
