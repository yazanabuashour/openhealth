DROP INDEX IF EXISTS idx_health_lab_result_canonical_slug_panel_id;
DROP INDEX IF EXISTS idx_health_lab_result_panel_id_display_order;

ALTER TABLE health_lab_result RENAME TO health_lab_result_old;

CREATE TABLE health_lab_result (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  panel_id INTEGER NOT NULL REFERENCES health_lab_panel(id) ON DELETE CASCADE,
  test_name TEXT NOT NULL,
  canonical_slug TEXT,
  value_text TEXT NOT NULL,
  value_numeric REAL,
  units TEXT,
  range_text TEXT,
  flag TEXT,
  display_order INTEGER NOT NULL
);

INSERT INTO health_lab_result (
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
)
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
FROM health_lab_result_old;

DROP TABLE health_lab_result_old;

CREATE INDEX idx_health_lab_result_canonical_slug_panel_id
  ON health_lab_result (canonical_slug, panel_id);

CREATE INDEX idx_health_lab_result_panel_id_display_order
  ON health_lab_result (panel_id, display_order);
