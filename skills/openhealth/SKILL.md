---
name: openhealth
description: Use this skill for local-first OpenHealth weight, body-composition, blood-pressure, medication, lab, or imaging data. Reject directly without tools or file reads for short dates without a year, year-first slash dates, invalid values, unsupported units/statuses, invalid lab slugs including slashes, empty optional text or notes, unsafe corrections/deletes, medication end before start, systolic not greater than diastolic, or any mixed-domain request with an invalid write; do not run validate. For valid requests, use this skill's runner contract, assume openhealth is on PATH, and pipe JSON directly to the installed runner; never command -v, --help, search source/docs/module cache/SQLite, or inspect code. Batch same-domain rows in one runner call; after writes, answer from returned entries unless a different final filter is requested.
license: MIT
compatibility: Requires local filesystem access and an installed openhealth binary on PATH.
---

# OpenHealth

Use the installed `openhealth` JSON runners:

- `openhealth weight`
- `openhealth body-composition`
- `openhealth blood-pressure`
- `openhealth medications`
- `openhealth labs`
- `openhealth imaging`

The command syntax above is complete. The configured local data path is already
available through the environment; do not call `--help`, inspect source, or add
`-db` unless the user explicitly provides a database path.
Assume `openhealth` is already installed on `PATH`; do not run `command -v
openhealth` before using it.
Do not inspect environment variables or search for database files before routine
runner calls; the runner already receives the configured local data path.

## Reject Before Tools

For the cases below, reject directly without running code, inspecting files,
searching the repo, checking the database, using the OpenHealth runner, or calling
the CLI when the request has:

| Issue | Response |
| --- | --- |
| ambiguous short date without year context, like `03/29` | ask for the year |
| year-first slash date, like `2026/03/31` | require `YYYY-MM-DD`; do not normalize |
| non-positive weight, systolic, diastolic, or pulse | reject as invalid |
| systolic not greater than diastolic | reject as invalid |
| unsupported weight unit, like `stone` | reject; pounds only |
| invalid body-composition percentage or contextual weight | reject as invalid |
| incomplete body-composition weight pair | reject as invalid |
| invalid lab analyte slug shape, including slashes, punctuation-only text, or empty text | reject as invalid |
| unsupported medication status | reject as invalid |
| empty optional text field or note string | reject as invalid |
| medication end date before start date | reject as invalid |
| mixed-domain request where any requested write is invalid | reject the whole request; do not partially write valid rows |

These are final-answer-only cases. Do not run a runner `validate` action for
them; answer from these rules.

Full slash dates with a year, like `03/29/2026` or `02/01/2026`, may be
normalized to `YYYY-MM-DD`. Short dates without a year still require a year
clarification.
Optional text fields cannot be cleared with runner JSON in this release; omit
them on corrections to preserve existing text, or provide a non-empty replacement
value.
Preserve narrative context in the narrowest matching note field. Use weight or
blood-pressure `note` for row context, medication `note` for course context, lab
collection `note` for collection or doctor-note context, lab `results[].notes`
for result-level notes, body-composition `note` for record context, imaging
`note` for record-level import or clinician context, and imaging `notes` for
result/report narrative. Note arrays preserve order and multiline text.

## Runner Contract

Pipe one JSON request to one runner and answer only from JSON `entries`,
`writes`, or `rejection_reason`. Run mixed requests as one call per domain.
Runner `entries` are already newest-first. Valid requests are validated before
database access.
Use one runner call for each valid domain operation. Batch multiple same-domain
rows into one JSON request, such as two blood-pressure readings in one
`record_blood_pressure` request. Write actions return `entries`; for write and
report tasks, answer from those returned `entries` unless the requested final
answer needs a different filter than the write response provides. Do not make a
follow-up list call just to confirm a write that already returned enough
matching `entries`.

Common write-and-report tasks should stay minimal:

- Record two blood-pressure readings, then report newest-first: one
  `record_blood_pressure` call; answer from returned `entries`.
- Record a medication and a lab, then report the active medication and latest
  lab: one `record_medications` call and one `record_labs` call; answer from
  returned `entries` when those are the rows just recorded.
- Correct the latest weight and blood pressure after a prior latest read: one
  `upsert_weights` call and one `correct_blood_pressure` call; answer from
  returned `entries`.

When a task writes data and then asks for a filtered list, make the final answer
match the filtered list response. Do not mention entries outside the requested
final filter unless the user explicitly asks for them. Omitted entries do not
need explanatory notes.
Use the `YYYY-MM-DD` dates returned by runner `entries` in list answers so
bounded ranges are explicit and machine-checkable.

Weights:

```json
{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb","note":"morning scale after run"}]}
{"action":"list_weights","list_mode":"latest"}
{"action":"list_weights","list_mode":"history","limit":2}
{"action":"list_weights","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Request JSON fields are `action`, `weights`, `list_mode`, `from_date`,
`to_date`, and `limit`. Each weight has `date`, `value`, `unit`, and optional
`note` for row-level context.
Use `upsert_weights` to write, reapply, or correct weights. Repeating a
same-date value is idempotent. A same-date different value updates the row.
Accepted units are `lb`, `lbs`, `pound`, and `pounds`; the runner normalizes them
to `lb`. For "two most recent" or any count greater than one, use `history`
with `limit`; `latest` returns one row.

Keep scale weight in `openhealth weight`. If a source row contains both scale
weight and body-fat percentage, call `openhealth weight` for the weight and
`openhealth body-composition` for the body-fat/body-composition data.

Body composition:

```json
{"action":"record_body_composition","records":[{"date":"2026-03-29","body_fat_percent":18.7,"weight_value":154.2,"weight_unit":"lb","method":"smart scale","note":"same row as weight import"}]}
{"action":"correct_body_composition","target":{"id":123},"record":{"date":"2026-03-29","body_fat_percent":18.1,"method":"DEXA"}}
{"action":"delete_body_composition","target":{"id":123}}
{"action":"list_body_composition","list_mode":"latest"}
{"action":"list_body_composition","list_mode":"history","limit":2}
{"action":"list_body_composition","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Request JSON fields are `action`, `records`, `record`, `target`, `list_mode`,
`from_date`, `to_date`, and `limit`. Each record has `date`, optional
`body_fat_percent`, optional contextual `weight_value`, optional `weight_unit`,
optional `method`, and optional `note`. At least one measurement is required.
`body_fat_percent` must be greater than 0 and less than or equal to 100.
Contextual `weight_value` must be greater than 0, `weight_unit` must normalize
to `lb`, and `weight_value` plus `weight_unit` must be provided together.
Repeating an exact record is idempotent and returns `already_exists`; a distinct
same-date record is stored as a separate record. Use corrections and deletes
with a target by `id`, or by `date` only when exactly one same-date record
exists. If a date is ambiguous, list body composition first and use the returned
record `id`.

Blood pressure:

```json
{"action":"record_blood_pressure","readings":[{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64,"note":"home cuff, seated"}]}
{"action":"correct_blood_pressure","readings":[{"date":"2026-03-29","systolic":121,"diastolic":77}]}
{"action":"list_blood_pressure","list_mode":"latest"}
{"action":"list_blood_pressure","list_mode":"history","limit":2}
{"action":"list_blood_pressure","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
```

Request JSON fields are `action`, `readings`, `list_mode`, `from_date`,
`to_date`, and `limit`. Each reading has `date`, `systolic`, `diastolic`,
optional positive integer `pulse`, and optional `note` for row-level context. Use
`record_blood_pressure` to add readings.
Use `correct_blood_pressure` when the user asks to correct an existing
same-date reading. Correction updates exactly one existing reading on the
requested date; if no same-date reading or multiple same-date readings exist,
The runner returns `rejected` with `rejection_reason`. For "two most recent" or
any count greater than one, use `history` with `limit`; `latest` returns one row.

Medications:

```json
{"action":"record_medications","medications":[{"name":"Levothyroxine","dosage_text":"25 mcg","start_date":"2026-01-01"}]}
{"action":"correct_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"},"medication":{"name":"Levothyroxine","dosage_text":"50 mcg","start_date":"2026-01-01","end_date":"2026-04-01"}}
{"action":"delete_medication","target":{"name":"Levothyroxine","start_date":"2026-01-01"}}
{"action":"list_medications","status":"active"}
{"action":"list_medications","status":"all"}
```

Request JSON fields are `action`, `medications`, `medication`, `target`, and
`status`. Each medication request has `name`, optional `dosage_text`,
`start_date`, optional `end_date`, and optional `note`. Use `record_medications`
with one or more courses. Repeating an exact same `name` and `start_date` course
is idempotent and returns `already_exists`; the same `name` and `start_date`
with different details is rejected and should be corrected with
`correct_medication`.
When the user provides dose details, put the full amount, route, form, frequency,
and delivery details in `dosage_text`, such as `2.5 mg subcutaneous injection
weekly`, `topical cream twice daily`, or `1 patch every 24 hours`.
Use `note` for extra medication-course narrative that does not belong in the
structured dose text, such as why the course changed or context from an import
file.
Use `correct_medication` or `delete_medication` with a target by `id`, or by
exact normalized `name` and `start_date`. The target must match exactly one
medication; zero or multiple matches return `rejected` with `rejection_reason`.
`active` is the default status; only `active` and `all` are supported.
For `status:"active"` answers, mention only active `entries`; never mention
inactive or ended courses, even to explain why they were omitted.
If a write response includes inactive or ended medication courses and the user
asked for active medications, run `list_medications` with `status:"active"` and
answer only from that filtered response.
After `record_medications`, do not call `list_medications` just to report the
active course that was returned by the write response.

Labs:

```json
{"action":"record_labs","collections":[{"date":"2026-03-29","note":"labs look stable, keep moving","panels":[{"panel_name":"Metabolic","results":[{"test_name":"Glucose","canonical_slug":"glucose","value_text":"89","value_numeric":89,"units":"mg/dL","range_text":"70-99","notes":["HIV 4th gen narrative","A1C context"]}]}]}]}
{"action":"correct_labs","target":{"date":"2026-03-29"},"collection":{"date":"2026-03-29","panels":[{"panel_name":"Thyroid","results":[{"test_name":"TSH","canonical_slug":"tsh","value_text":"3.1","value_numeric":3.1,"units":"uIU/mL"}]}]}}
{"action":"patch_labs","target":{"id":123},"result_updates":[{"panel_name":"Metabolic","match":{"canonical_slug":"glucose"},"result":{"test_name":"Glucose","canonical_slug":"glucose","value_text":"92","value_numeric":92,"units":"mg/dL","notes":["corrected value note"]}}]}
{"action":"delete_labs","target":{"date":"2026-03-29"}}
{"action":"list_labs","list_mode":"latest"}
{"action":"list_labs","list_mode":"history","limit":2}
{"action":"list_labs","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
{"action":"list_labs","list_mode":"latest","analyte_slug":"glucose"}
```

Request JSON fields are `action`, `collections`, `collection`,
`result_updates`, `target`, `list_mode`, `from_date`, `to_date`, `limit`, and
`analyte_slug`. Each collection request has `date`, optional `note`, nested
`panels`, and nested `results`. Each panel has `panel_name` and `results`. Each
result has `test_name`, optional `canonical_slug`, `value_text`, optional
`value_numeric`, optional `units`, optional `range_text`, optional `flag`, and
optional ordered `notes` for result-level narrative. Put lab result notes in
`results[].notes`; do not merge them into `value_text`, `flag`, or collection
`note`.

Use lowercase kebab-case `canonical_slug` and `analyte_slug` values derived from
the lab or test name, such as `vitamin-d`, `hemoglobin-a1c`, `ferritin`, and
`urine-albumin-creatinine-ratio`. The runner normalizes spaces and underscores
to hyphens and rejects empty or non-kebab-case slug shapes.
Use `record_labs` with one or more date-only collections. Repeating an exact
same-date collection is idempotent and returns `already_exists`; a different
same-date collection is stored as a separate collection. Use `correct_labs` for
full collection replacement: the replacement collection should contain only the
panels/results the user wants stored after correction. Do not preserve sibling
panels or results during `correct_labs` unless the user explicitly asks to keep
them. Use `patch_labs` for one-result corrections that should preserve sibling
panels and results. A `patch_labs` update requires
`panel_name`, a `match` with exactly one of `canonical_slug` or `test_name`, and
a full replacement `result`. Use `correct_labs`, `patch_labs`, or `delete_labs`
with a target by `id`, or by `date` only when exactly one collection exists on
that date. Zero or multiple target matches return `rejected` with
`rejection_reason`; if a date is ambiguous, list labs first and use the returned
collection `id`.

For "two most recent" or any count greater than one, use `history` with
`limit`; `latest` returns one matching collection. `analyte_slug` filters nested
results to that canonical analyte and omits collections without matching
results.
After `record_labs`, do not call `list_labs` just to report the latest
collection or analyte result that was returned by the write response.

Imaging:

```json
{"action":"record_imaging","records":[{"date":"2026-03-29","modality":"X-ray","body_site":"chest","title":"Chest X-ray","summary":"No acute cardiopulmonary abnormality.","impression":"Normal chest radiograph.","note":"ordered for cough","notes":["XR TOE RIGHT narrative","US Head/Neck findings"]}]}
{"action":"correct_imaging","target":{"id":123},"record":{"date":"2026-03-29","modality":"CT","body_site":"chest","summary":"Stable small pulmonary nodule.","note":"follow-up scan"}}
{"action":"delete_imaging","target":{"id":123}}
{"action":"list_imaging","list_mode":"latest"}
{"action":"list_imaging","list_mode":"history","limit":2}
{"action":"list_imaging","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
{"action":"list_imaging","list_mode":"history","modality":"x-ray","body_site":"CHEST"}
```

Request JSON fields are `action`, `records`, `record`, `target`, `list_mode`,
`from_date`, `to_date`, `limit`, optional `modality`, and optional `body_site`.
Each imaging record has `date`, `modality`, optional `body_site`, optional
`title`, required `summary`, optional `impression`, optional `note`, and optional
ordered `notes`.
`date`, `modality`, and `summary` are required. Use `summary` for the main scan
or report summary, `impression` for the radiology impression when present, and
`note` for extra import-file or clinician narrative. Put imaging result/report
narrative in `notes` so it stays attached to the imaging record without
overwriting `summary` or `impression`. `modality` and `body_site` filters are
exact case-insensitive matches after trimming. Repeating an exact record is
idempotent and returns `already_exists`; a distinct same-date record is stored as
a separate record. Use corrections and deletes with a target by `id`, or by
`date` only when exactly one same-date imaging record exists. If a date is
ambiguous, list imaging first and use the returned record `id`.

Do not run repo-wide file discovery or broad searches for routine user-data
tasks; do not inspect source files, module-cache docs, or SQLite directly.
