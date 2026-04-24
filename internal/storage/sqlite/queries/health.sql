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

-- name: FindManualWeightEntry :one
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
  AND source = 'manual'
  AND substr(recorded_at, 1, 10) = sqlc.arg('recorded_date')
  AND unit = sqlc.arg('unit')
ORDER BY id DESC
LIMIT 1;

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
  note = COALESCE(sqlc.narg('note'), note),
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
  note,
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

-- name: CreateBloodPressureEntry :one
INSERT INTO health_blood_pressure_entry (
  recorded_at,
  systolic,
  diastolic,
  pulse,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at
) VALUES (
  sqlc.arg('recorded_at'),
  sqlc.arg('systolic'),
  sqlc.arg('diastolic'),
  sqlc.narg('pulse'),
  sqlc.narg('note'),
  sqlc.arg('source'),
  sqlc.arg('source_record_hash'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at')
)
RETURNING
  id,
  recorded_at,
  systolic,
  diastolic,
  pulse,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: UpdateBloodPressureEntry :one
UPDATE health_blood_pressure_entry
SET
  recorded_at = sqlc.arg('recorded_at'),
  systolic = sqlc.arg('systolic'),
  diastolic = sqlc.arg('diastolic'),
  pulse = sqlc.narg('pulse'),
  note = sqlc.narg('note'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING
  id,
  recorded_at,
  systolic,
  diastolic,
  pulse,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: DeleteBloodPressureEntry :one
UPDATE health_blood_pressure_entry
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id;

-- name: ListMedicationCourses :many
SELECT
  id,
  name,
  dosage_text,
  start_date,
  end_date,
  note,
  source,
  created_at,
  updated_at,
  deleted_at
FROM health_medication_course
WHERE deleted_at IS NULL
ORDER BY start_date DESC, id DESC;

-- name: CreateMedicationCourse :one
INSERT INTO health_medication_course (
  name,
  dosage_text,
  start_date,
  end_date,
  note,
  source,
  created_at,
  updated_at
) VALUES (
  sqlc.arg('name'),
  sqlc.narg('dosage_text'),
  sqlc.arg('start_date'),
  sqlc.narg('end_date'),
  sqlc.narg('note'),
  sqlc.arg('source'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at')
)
RETURNING
  id,
  name,
  dosage_text,
  start_date,
  end_date,
  note,
  source,
  created_at,
  updated_at,
  deleted_at;

-- name: UpdateMedicationCourse :one
UPDATE health_medication_course
SET
  name = sqlc.arg('name'),
  dosage_text = sqlc.narg('dosage_text'),
  start_date = sqlc.arg('start_date'),
  end_date = sqlc.narg('end_date'),
  note = sqlc.narg('note'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING
  id,
  name,
  dosage_text,
  start_date,
  end_date,
  note,
  source,
  created_at,
  updated_at,
  deleted_at;

-- name: DeleteMedicationCourse :one
UPDATE health_medication_course
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id;

-- name: ListActiveMedicationCourses :many
SELECT
  id,
  name,
  dosage_text,
  start_date,
  end_date,
  note,
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
  note,
  source,
  created_at,
  updated_at,
  deleted_at
FROM health_lab_collection
WHERE deleted_at IS NULL
ORDER BY collected_at DESC, id DESC;

-- name: GetLabCollection :one
SELECT
  id,
  collected_at,
  note,
  source,
  created_at,
  updated_at,
  deleted_at
FROM health_lab_collection
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL;

-- name: CreateLabCollection :one
INSERT INTO health_lab_collection (
  collected_at,
  note,
  source,
  created_at,
  updated_at
) VALUES (
  sqlc.arg('collected_at'),
  sqlc.narg('note'),
  sqlc.arg('source'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at')
)
RETURNING
  id,
  collected_at,
  note,
  source,
  created_at,
  updated_at,
  deleted_at;

-- name: UpdateLabCollection :one
UPDATE health_lab_collection
SET
  collected_at = sqlc.arg('collected_at'),
  note = sqlc.narg('note'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING
  id,
  collected_at,
  note,
  source,
  created_at,
  updated_at,
  deleted_at;

-- name: DeleteLabCollection :one
UPDATE health_lab_collection
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id;

-- name: ListLabPanels :many
SELECT
  id,
  collection_id,
  panel_name,
  display_order
FROM health_lab_panel
ORDER BY collection_id ASC, display_order ASC, id ASC;

-- name: DeleteLabPanelsByCollection :exec
DELETE FROM health_lab_panel
WHERE collection_id = sqlc.arg('collection_id');

-- name: CreateLabPanel :one
INSERT INTO health_lab_panel (
  collection_id,
  panel_name,
  display_order
) VALUES (
  sqlc.arg('collection_id'),
  sqlc.arg('panel_name'),
  sqlc.arg('display_order')
)
RETURNING
  id,
  collection_id,
  panel_name,
  display_order;

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

-- name: ListLabResultNotes :many
SELECT
  id,
  lab_result_id,
  note_text,
  display_order
FROM health_lab_result_note
ORDER BY lab_result_id ASC, display_order ASC, id ASC;

-- name: CreateLabResult :one
INSERT INTO health_lab_result (
  panel_id,
  test_name,
  canonical_slug,
  value_text,
  value_numeric,
  units,
  range_text,
  flag,
  display_order
) VALUES (
  sqlc.arg('panel_id'),
  sqlc.arg('test_name'),
  sqlc.narg('canonical_slug'),
  sqlc.arg('value_text'),
  sqlc.narg('value_numeric'),
  sqlc.narg('units'),
  sqlc.narg('range_text'),
  sqlc.narg('flag'),
  sqlc.arg('display_order')
)
RETURNING
  id,
  panel_id,
  test_name,
  canonical_slug,
  value_text,
  value_numeric,
  units,
  range_text,
  flag,
  display_order;

-- name: CreateLabResultNote :one
INSERT INTO health_lab_result_note (
  lab_result_id,
  note_text,
  display_order
) VALUES (
  sqlc.arg('lab_result_id'),
  sqlc.arg('note_text'),
  sqlc.arg('display_order')
)
RETURNING
  id,
  lab_result_id,
  note_text,
  display_order;

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
  AND health_lab_collection.deleted_at IS NULL
ORDER BY health_lab_collection.collected_at DESC, health_lab_result.id DESC;

-- name: ListBodyCompositionEntries :many
SELECT
  id,
  recorded_at,
  body_fat_percent,
  weight_value,
  weight_unit,
  method,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at
FROM health_body_composition_entry
WHERE deleted_at IS NULL
  AND (sqlc.narg('from_recorded_at') IS NULL OR recorded_at >= sqlc.narg('from_recorded_at'))
  AND (sqlc.narg('to_recorded_at') IS NULL OR recorded_at <= sqlc.narg('to_recorded_at'))
ORDER BY recorded_at DESC, id DESC
LIMIT CASE
  WHEN sqlc.narg('limit_count') IS NULL THEN -1
  ELSE sqlc.narg('limit_count')
END;

-- name: CreateBodyCompositionEntry :one
INSERT INTO health_body_composition_entry (
  recorded_at,
  body_fat_percent,
  weight_value,
  weight_unit,
  method,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at
) VALUES (
  sqlc.arg('recorded_at'),
  sqlc.narg('body_fat_percent'),
  sqlc.narg('weight_value'),
  sqlc.narg('weight_unit'),
  sqlc.narg('method'),
  sqlc.narg('note'),
  sqlc.arg('source'),
  sqlc.arg('source_record_hash'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at')
)
RETURNING
  id,
  recorded_at,
  body_fat_percent,
  weight_value,
  weight_unit,
  method,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: UpdateBodyCompositionEntry :one
UPDATE health_body_composition_entry
SET
  recorded_at = sqlc.arg('recorded_at'),
  body_fat_percent = sqlc.narg('body_fat_percent'),
  weight_value = sqlc.narg('weight_value'),
  weight_unit = sqlc.narg('weight_unit'),
  method = sqlc.narg('method'),
  note = sqlc.narg('note'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING
  id,
  recorded_at,
  body_fat_percent,
  weight_value,
  weight_unit,
  method,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: DeleteBodyCompositionEntry :one
UPDATE health_body_composition_entry
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id;

-- name: ListSleepEntries :many
SELECT
  id,
  recorded_at,
  quality_score,
  wakeup_count,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at
FROM health_sleep_entry
WHERE deleted_at IS NULL
  AND (sqlc.narg('from_recorded_at') IS NULL OR recorded_at >= sqlc.narg('from_recorded_at'))
  AND (sqlc.narg('to_recorded_at') IS NULL OR recorded_at <= sqlc.narg('to_recorded_at'))
ORDER BY recorded_at DESC, id DESC
LIMIT CASE
  WHEN sqlc.narg('limit_count') IS NULL THEN -1
  ELSE sqlc.narg('limit_count')
END;

-- name: FindManualSleepEntry :one
SELECT
  id,
  recorded_at,
  quality_score,
  wakeup_count,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at
FROM health_sleep_entry
WHERE deleted_at IS NULL
  AND source = 'manual'
  AND substr(recorded_at, 1, 10) = sqlc.arg('recorded_date')
ORDER BY id DESC
LIMIT 1;

-- name: CreateSleepEntry :one
INSERT INTO health_sleep_entry (
  recorded_at,
  quality_score,
  wakeup_count,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at
) VALUES (
  sqlc.arg('recorded_at'),
  sqlc.arg('quality_score'),
  sqlc.narg('wakeup_count'),
  sqlc.narg('note'),
  sqlc.arg('source'),
  sqlc.arg('source_record_hash'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at')
)
RETURNING
  id,
  recorded_at,
  quality_score,
  wakeup_count,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: UpdateSleepEntry :one
UPDATE health_sleep_entry
SET
  recorded_at = COALESCE(sqlc.narg('recorded_at'), recorded_at),
  quality_score = COALESCE(sqlc.narg('quality_score'), quality_score),
  wakeup_count = COALESCE(sqlc.narg('wakeup_count'), wakeup_count),
  note = COALESCE(sqlc.narg('note'), note),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING
  id,
  recorded_at,
  quality_score,
  wakeup_count,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: DeleteSleepEntry :one
UPDATE health_sleep_entry
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id;

-- name: ListImagingRecords :many
SELECT
  id,
  performed_at,
  modality,
  body_site,
  title,
  summary,
  impression,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at
FROM health_imaging_record
WHERE deleted_at IS NULL
  AND (sqlc.narg('from_performed_at') IS NULL OR performed_at >= sqlc.narg('from_performed_at'))
  AND (sqlc.narg('to_performed_at') IS NULL OR performed_at <= sqlc.narg('to_performed_at'))
  AND (sqlc.narg('modality') IS NULL OR lower(modality) = lower(sqlc.narg('modality')))
  AND (sqlc.narg('body_site') IS NULL OR (body_site IS NOT NULL AND lower(body_site) = lower(sqlc.narg('body_site'))))
ORDER BY performed_at DESC, id DESC
LIMIT CASE
  WHEN sqlc.narg('limit_count') IS NULL THEN -1
  ELSE sqlc.narg('limit_count')
END;

-- name: ListImagingRecordNotes :many
SELECT
  id,
  imaging_record_id,
  note_text,
  display_order
FROM health_imaging_record_note
ORDER BY imaging_record_id ASC, display_order ASC, id ASC;

-- name: CreateImagingRecord :one
INSERT INTO health_imaging_record (
  performed_at,
  modality,
  body_site,
  title,
  summary,
  impression,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at
) VALUES (
  sqlc.arg('performed_at'),
  sqlc.arg('modality'),
  sqlc.narg('body_site'),
  sqlc.narg('title'),
  sqlc.arg('summary'),
  sqlc.narg('impression'),
  sqlc.narg('note'),
  sqlc.arg('source'),
  sqlc.arg('source_record_hash'),
  sqlc.arg('created_at'),
  sqlc.arg('updated_at')
)
RETURNING
  id,
  performed_at,
  modality,
  body_site,
  title,
  summary,
  impression,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: CreateImagingRecordNote :one
INSERT INTO health_imaging_record_note (
  imaging_record_id,
  note_text,
  display_order
) VALUES (
  sqlc.arg('imaging_record_id'),
  sqlc.arg('note_text'),
  sqlc.arg('display_order')
)
RETURNING
  id,
  imaging_record_id,
  note_text,
  display_order;

-- name: DeleteImagingRecordNotesByRecord :exec
DELETE FROM health_imaging_record_note
WHERE imaging_record_id = sqlc.arg('imaging_record_id');

-- name: UpdateImagingRecord :one
UPDATE health_imaging_record
SET
  performed_at = sqlc.arg('performed_at'),
  modality = sqlc.arg('modality'),
  body_site = sqlc.narg('body_site'),
  title = sqlc.narg('title'),
  summary = sqlc.arg('summary'),
  impression = sqlc.narg('impression'),
  note = sqlc.narg('note'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING
  id,
  performed_at,
  modality,
  body_site,
  title,
  summary,
  impression,
  note,
  source,
  source_record_hash,
  created_at,
  updated_at,
  deleted_at;

-- name: DeleteImagingRecord :one
UPDATE health_imaging_record
SET
  deleted_at = sqlc.arg('deleted_at'),
  updated_at = sqlc.arg('updated_at')
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING id;

-- name: GetConfigValue :one
SELECT
  key,
  value_json,
  updated_at
FROM openhealth_config
WHERE key = sqlc.arg('key');

-- name: ListConfigValues :many
SELECT
  key,
  value_json,
  updated_at
FROM openhealth_config
ORDER BY key ASC;

-- name: UpsertConfigValue :one
INSERT INTO openhealth_config (
  key,
  value_json,
  updated_at
) VALUES (
  sqlc.arg('key'),
  sqlc.arg('value_json'),
  sqlc.arg('updated_at')
)
ON CONFLICT(key) DO UPDATE SET
  value_json = excluded.value_json,
  updated_at = excluded.updated_at
RETURNING
  key,
  value_json,
  updated_at;

-- name: DeleteConfigValue :execrows
DELETE FROM openhealth_config
WHERE key = sqlc.arg('key');
