-- name: ListWeightEntries :many
SELECT
  id,
  recorded_at,
  value,
  unit,
  source,
  source_record_hash,
  note,
  created_at,
  updated_at,
  deleted_at
FROM health_weight_entry
WHERE deleted_at IS NULL
  AND (sqlc.narg('from_recorded_at') IS NULL OR recorded_at >= sqlc.narg('from_recorded_at'))
  AND (sqlc.narg('to_recorded_at') IS NULL OR recorded_at <= sqlc.narg('to_recorded_at'))
ORDER BY recorded_at DESC, id DESC
LIMIT CASE
  WHEN sqlc.narg('limit_count') IS NULL THEN -1
  ELSE sqlc.narg('limit_count')
END;

-- name: CreateWeightEntry :one
INSERT INTO health_weight_entry (
  recorded_at,
  value,
  unit,
  source,
  source_record_hash,
  note,
  created_at,
  updated_at
) VALUES (
  sqlc.arg('recorded_at'),
  sqlc.arg('value'),
  sqlc.arg('unit'),
  sqlc.arg('source'),
  sqlc.arg('source_record_hash'),
  sqlc.narg('note'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at')
)
RETURNING
  id,
  recorded_at,
  value,
  unit,
  source,
  source_record_hash,
  note,
  created_at,
  updated_at,
  deleted_at;

-- name: UpdateWeightEntry :one
UPDATE health_weight_entry
SET
  recorded_at = COALESCE(sqlc.narg('recorded_at'), recorded_at),
  value = COALESCE(sqlc.narg('value'), value),
  unit = COALESCE(sqlc.narg('unit'), unit),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING
  id,
  recorded_at,
  value,
  unit,
  source,
  source_record_hash,
  note,
  created_at,
  updated_at,
  deleted_at;

-- name: DeleteWeightEntry :one
UPDATE health_weight_entry
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id;

-- name: ListBloodPressureEntries :many
SELECT
  id,
  recorded_at,
  systolic,
  diastolic,
  pulse,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at
FROM health_blood_pressure_entry
WHERE deleted_at IS NULL
  AND (sqlc.narg('from_recorded_at') IS NULL OR recorded_at >= sqlc.narg('from_recorded_at'))
  AND (sqlc.narg('to_recorded_at') IS NULL OR recorded_at <= sqlc.narg('to_recorded_at'))
ORDER BY recorded_at DESC, id DESC
LIMIT CASE
  WHEN sqlc.narg('limit_count') IS NULL THEN -1
  ELSE sqlc.narg('limit_count')
END;

-- name: ListMedicationCourses :many
SELECT
  id,
  name,
  dosage_text,
  start_date,
  end_date,
  source,
  created_at,
  updated_at,
  deleted_at
FROM health_medication_course
WHERE deleted_at IS NULL
ORDER BY start_date DESC, id DESC;

-- name: ListActiveMedicationCourses :many
SELECT
  id,
  name,
  dosage_text,
  start_date,
  end_date,
  source,
  created_at,
  updated_at,
  deleted_at
FROM health_medication_course
WHERE deleted_at IS NULL
  AND (end_date IS NULL OR end_date >= sqlc.arg('today'))
ORDER BY start_date DESC, id DESC;

-- name: CountActiveMedicationCourses :one
SELECT COUNT(*) AS count
FROM health_medication_course
WHERE deleted_at IS NULL
  AND (end_date IS NULL OR end_date >= sqlc.arg('today'));

-- name: ListLabCollections :many
SELECT
  id,
  collected_at,
  source,
  created_at
FROM health_lab_collection
ORDER BY collected_at DESC, id DESC;

-- name: ListLabPanels :many
SELECT
  id,
  collection_id,
  panel_name,
  display_order
FROM health_lab_panel
ORDER BY collection_id ASC, display_order ASC, id ASC;

-- name: ListLabResults :many
SELECT
  id,
  panel_id,
  test_name,
  canonical_slug,
  value_text,
  value_numeric,
  units,
  range_text,
  flag,
  display_order
FROM health_lab_result
ORDER BY panel_id ASC, display_order ASC, id ASC;

-- name: ListLabResultsWithCollection :many
SELECT
  health_lab_result.id,
  health_lab_result.panel_id,
  health_lab_result.test_name,
  health_lab_result.canonical_slug,
  health_lab_result.value_text,
  health_lab_result.value_numeric,
  health_lab_result.units,
  health_lab_result.range_text,
  health_lab_result.flag,
  health_lab_result.display_order,
  health_lab_collection.collected_at,
  health_lab_collection.id AS collection_id,
  health_lab_panel.panel_name
FROM health_lab_result
INNER JOIN health_lab_panel ON health_lab_panel.id = health_lab_result.panel_id
INNER JOIN health_lab_collection ON health_lab_collection.id = health_lab_panel.collection_id
WHERE health_lab_result.canonical_slug IS NOT NULL
ORDER BY health_lab_collection.collected_at DESC, health_lab_result.id DESC;
