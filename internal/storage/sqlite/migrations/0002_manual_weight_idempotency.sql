CREATE UNIQUE INDEX idx_health_weight_entry_manual_date_unit_unique
  ON health_weight_entry (substr(recorded_at, 1, 10), unit)
  WHERE deleted_at IS NULL
    AND source = 'manual';
