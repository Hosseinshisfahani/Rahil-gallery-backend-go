DROP INDEX IF EXISTS idx_customer_profiles_crm_purchased_categories;
DROP INDEX IF EXISTS idx_customer_profiles_crm_marriage_date;
DROP INDEX IF EXISTS idx_customer_profiles_crm_birthday;
DROP INDEX IF EXISTS idx_customer_profiles_crm_first_visit;
DROP INDEX IF EXISTS idx_customer_profiles_crm_customer_type;
DROP INDEX IF EXISTS idx_customer_profiles_crm_gender;
DROP INDEX IF EXISTS idx_customer_profiles_crm_age_range;

ALTER TABLE customer_profiles
    DROP COLUMN IF EXISTS crm_marriage_date,
    DROP COLUMN IF EXISTS crm_birthday,
    DROP COLUMN IF EXISTS crm_first_visit_date,
    DROP COLUMN IF EXISTS crm_customer_type,
    DROP COLUMN IF EXISTS crm_gender,
    DROP COLUMN IF EXISTS crm_age_range;
