-- CRM advanced search: indexed profile fields extracted from import_profile JSONB.

ALTER TABLE customer_profiles
    ADD COLUMN IF NOT EXISTS crm_age_range TEXT
        GENERATED ALWAYS AS (NULLIF(import_profile->>'customerAgeRange', '')) STORED,
    ADD COLUMN IF NOT EXISTS crm_gender TEXT
        GENERATED ALWAYS AS (NULLIF(import_profile->>'gender', '')) STORED,
    ADD COLUMN IF NOT EXISTS crm_customer_type TEXT
        GENERATED ALWAYS AS (NULLIF(import_profile->>'customerType', '')) STORED,
    ADD COLUMN IF NOT EXISTS crm_first_visit_date TEXT
        GENERATED ALWAYS AS (NULLIF(import_profile->>'firstVisitDate', '')) STORED,
    ADD COLUMN IF NOT EXISTS crm_birthday TEXT
        GENERATED ALWAYS AS (NULLIF(import_profile->>'birthday', '')) STORED,
    ADD COLUMN IF NOT EXISTS crm_marriage_date TEXT
        GENERATED ALWAYS AS (NULLIF(import_profile->>'marriageDate', '')) STORED;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_crm_age_range
    ON customer_profiles (crm_age_range)
    WHERE crm_age_range IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_crm_gender
    ON customer_profiles (crm_gender)
    WHERE crm_gender IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_crm_customer_type
    ON customer_profiles (crm_customer_type)
    WHERE crm_customer_type IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_crm_first_visit
    ON customer_profiles (crm_first_visit_date)
    WHERE crm_first_visit_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_crm_birthday
    ON customer_profiles (crm_birthday)
    WHERE crm_birthday IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_crm_marriage_date
    ON customer_profiles (crm_marriage_date)
    WHERE crm_marriage_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_crm_purchased_categories
    ON customer_profiles USING gin ((import_profile->'purchasedCategories'));
