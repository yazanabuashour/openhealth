CREATE TABLE health_sleep_entry (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recorded_at TEXT NOT NULL,
  quality_score INTEGER NOT NULL CHECK (quality_score >= 1 AND quality_score <= 5),
  wakeup_count INTEGER CHECK (wakeup_count IS NULL OR wakeup_count >= 0),
  note TEXT,
  source TEXT NOT NULL,
  source_record_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  deleted_at TEXT
);

CREATE INDEX idx_health_sleep_entry_recorded_at_desc
  ON health_sleep_entry (recorded_at DESC, id DESC);

CREATE UNIQUE INDEX idx_health_sleep_entry_manual_date_unique
  ON health_sleep_entry (substr(recorded_at, 1, 10))
  WHERE deleted_at IS NULL
    AND source = 'manual';
