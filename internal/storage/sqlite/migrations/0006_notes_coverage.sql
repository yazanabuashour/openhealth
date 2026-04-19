ALTER TABLE health_blood_pressure_entry ADD COLUMN note TEXT;

CREATE TABLE health_lab_result_note (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lab_result_id INTEGER NOT NULL REFERENCES health_lab_result(id) ON DELETE CASCADE,
  note_text TEXT NOT NULL,
  display_order INTEGER NOT NULL
);

CREATE INDEX idx_health_lab_result_note_result_id_display_order
  ON health_lab_result_note (lab_result_id, display_order, id);

CREATE TABLE health_imaging_record_note (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  imaging_record_id INTEGER NOT NULL REFERENCES health_imaging_record(id) ON DELETE CASCADE,
  note_text TEXT NOT NULL,
  display_order INTEGER NOT NULL
);

CREATE INDEX idx_health_imaging_record_note_record_id_display_order
  ON health_imaging_record_note (imaging_record_id, display_order, id);
