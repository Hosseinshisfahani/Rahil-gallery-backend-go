-- Persist last provider error for failed/partial bulk SMS jobs
ALTER TABLE sms_jobs ADD COLUMN last_error TEXT;
