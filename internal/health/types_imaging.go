package health

import "time"

type ImagingRecord struct {
	ID               int
	PerformedAt      time.Time
	Modality         string
	BodySite         *string
	Title            *string
	Summary          string
	Impression       *string
	Note             *string
	Notes            []string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type ImagingRecordInput struct {
	PerformedAt time.Time
	Modality    string
	BodySite    *string
	Title       *string
	Summary     string
	Impression  *string
	Note        *string
	Notes       []string
}

type ImagingListParams struct {
	HistoryFilter
	Modality *string
	BodySite *string
}

type CreateImagingRecordParams struct {
	PerformedAt      time.Time
	Modality         string
	BodySite         *string
	Title            *string
	Summary          string
	Impression       *string
	Note             *string
	Notes            []string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateImagingRecordParams struct {
	ID          int
	PerformedAt time.Time
	Modality    string
	BodySite    *string
	Title       *string
	Summary     string
	Impression  *string
	Note        *string
	Notes       []string
	UpdatedAt   time.Time
}

type DeleteImagingRecordParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}
