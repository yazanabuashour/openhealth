ALTER TABLE health_lab_collection ADD COLUMN note TEXT;

ALTER TABLE health_medication_course ADD COLUMN note TEXT;

CREATE TABLE health_body_composition_entry (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recorded_at TEXT NOT NULL,
  body_fat_percent REAL CHECK (body_fat_percent IS NULL OR (body_fat_percent > 0 AND body_fat_percent <= 100)),
  weight_value REAL CHECK (weight_value IS NULL OR weight_value > 0),
  weight_unit TEXT CHECK (weight_unit IS NULL OR weight_unit = 'lb'),
  method TEXT,
  note TEXT,
  source TEXT NOT NULL,
  source_record_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  deleted_at TEXT,
  CHECK (body_fat_percent IS NOT NULL OR weight_value IS NOT NULL),
  CHECK ((weight_value IS NULL AND weight_unit IS NULL) OR (weight_value IS NOT NULL AND weight_unit IS NOT NULL))
);

CREATE INDEX idx_health_body_composition_entry_recorded_at_desc
  ON health_body_composition_entry (recorded_at DESC, id DESC);

CREATE TABLE health_imaging_record (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  performed_at TEXT NOT NULL,
  modality TEXT NOT NULL,
  body_site TEXT,
  title TEXT,
  summary TEXT NOT NULL,
  impression TEXT,
  note TEXT,
  source TEXT NOT NULL,
  source_record_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  deleted_at TEXT
);

CREATE INDEX idx_health_imaging_record_performed_at_desc
  ON health_imaging_record (performed_at DESC, id DESC);
