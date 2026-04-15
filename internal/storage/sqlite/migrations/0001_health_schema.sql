CREATE TABLE health_weight_entry (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recorded_at TEXT NOT NULL,
  value REAL NOT NULL CHECK (value > 0),
  unit TEXT NOT NULL CHECK (unit = 'lb'),
  source TEXT NOT NULL,
  source_record_hash TEXT NOT NULL,
  note TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  deleted_at TEXT
);

CREATE INDEX idx_health_weight_entry_recorded_at_desc
  ON health_weight_entry (recorded_at DESC, id DESC);

CREATE TABLE health_blood_pressure_entry (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recorded_at TEXT NOT NULL,
  systolic INTEGER NOT NULL CHECK (systolic > 0),
  diastolic INTEGER NOT NULL CHECK (diastolic > 0),
  pulse INTEGER,
  source TEXT NOT NULL,
  source_record_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  deleted_at TEXT
);

CREATE INDEX idx_health_blood_pressure_entry_recorded_at_desc
  ON health_blood_pressure_entry (recorded_at DESC, id DESC);

CREATE TABLE health_medication_course (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  dosage_text TEXT,
  start_date TEXT NOT NULL,
  end_date TEXT,
  source TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  deleted_at TEXT
);

CREATE INDEX idx_health_medication_course_start_date_desc
  ON health_medication_course (start_date DESC, id DESC);

CREATE TABLE health_lab_collection (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  collected_at TEXT NOT NULL,
  source TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX idx_health_lab_collection_collected_at_desc
  ON health_lab_collection (collected_at DESC, id DESC);

CREATE TABLE health_lab_panel (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  collection_id INTEGER NOT NULL REFERENCES health_lab_collection(id) ON DELETE CASCADE,
  panel_name TEXT NOT NULL,
  display_order INTEGER NOT NULL
);

CREATE TABLE health_lab_result (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  panel_id INTEGER NOT NULL REFERENCES health_lab_panel(id) ON DELETE CASCADE,
  test_name TEXT NOT NULL,
  canonical_slug TEXT CHECK (
    canonical_slug IS NULL OR canonical_slug IN (
      'tsh',
      'free-t4',
      'cholesterol-total',
      'ldl',
      'hdl',
      'triglycerides',
      'glucose'
    )
  ),
  value_text TEXT NOT NULL,
  value_numeric REAL,
  units TEXT,
  range_text TEXT,
  flag TEXT,
  display_order INTEGER NOT NULL
);

CREATE INDEX idx_health_lab_result_canonical_slug_panel_id
  ON health_lab_result (canonical_slug, panel_id);

CREATE INDEX idx_health_lab_result_panel_id_display_order
  ON health_lab_result (panel_id, display_order);
