ALTER TABLE health_lab_collection
  ADD COLUMN updated_at TEXT;

UPDATE health_lab_collection
SET updated_at = created_at
WHERE updated_at IS NULL;

ALTER TABLE health_lab_collection
  ADD COLUMN deleted_at TEXT;
