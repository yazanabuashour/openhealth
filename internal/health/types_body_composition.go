package health

import "time"

type BodyCompositionEntry struct {
	ID               int
	RecordedAt       time.Time
	BodyFatPercent   *float64
	WeightValue      *float64
	WeightUnit       *WeightUnit
	Method           *string
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type BodyCompositionInput struct {
	RecordedAt     time.Time
	BodyFatPercent *float64
	WeightValue    *float64
	WeightUnit     *WeightUnit
	Method         *string
	Note           *string
}

type CreateBodyCompositionEntryParams struct {
	RecordedAt       time.Time
	BodyFatPercent   *float64
	WeightValue      *float64
	WeightUnit       *WeightUnit
	Method           *string
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateBodyCompositionEntryParams struct {
	ID             int
	RecordedAt     time.Time
	BodyFatPercent *float64
	WeightValue    *float64
	WeightUnit     *WeightUnit
	Method         *string
	Note           *string
	UpdatedAt      time.Time
}

type DeleteBodyCompositionEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}
