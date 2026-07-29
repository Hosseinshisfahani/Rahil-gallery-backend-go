ALTER TABLE customers
    DROP COLUMN IF EXISTS melli_code,
    DROP COLUMN IF EXISTS postal_code,
    DROP COLUMN IF EXISTS marketer_note;
