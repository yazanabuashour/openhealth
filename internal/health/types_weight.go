package health

import "time"

type WeightUnit string

const WeightUnitLb WeightUnit = "lb"

type WeightRange string

const (
	WeightRange30d WeightRange = "30d"
	WeightRange90d WeightRange = "90d"
	WeightRange1y  WeightRange = "1y"
	WeightRangeAll WeightRange = "all"
)

type WeightEntry struct {
	ID               int
	RecordedAt       time.Time
	Value            float64
	Unit             WeightUnit
	Source           string
	SourceRecordHash string
	Note             *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type WeightWriteStatus string

const (
	WeightWriteStatusCreated       WeightWriteStatus = "created"
	WeightWriteStatusAlreadyExists WeightWriteStatus = "already_exists"
	WeightWriteStatusUpdated       WeightWriteStatus = "updated"
)

type WeightWriteResult struct {
	Entry  WeightEntry
	Status WeightWriteStatus
}

type MovingAveragePoint struct {
	RecordedAt time.Time
	Value      float64
}

type MonthlyAverageBucket struct {
	Month string
	Value float64
}

type WeightTrend struct {
	Range                 WeightRange
	RawPoints             []WeightEntry
	MovingAveragePoints   []MovingAveragePoint
	MonthlyAverageBuckets []MonthlyAverageBucket
}

type WeightRecordInput struct {
	RecordedAt time.Time
	Value      float64
	Unit       WeightUnit
	Note       *string
}

type WeightUpdateInput struct {
	RecordedAt *time.Time
	Value      *float64
	Unit       *WeightUnit
	Note       *string
}

type WeightTrendParams struct {
	Range WeightRange
}

type CreateWeightEntryParams struct {
	RecordedAt       time.Time
	Value            float64
	Unit             WeightUnit
	Source           string
	SourceRecordHash string
	Note             *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateWeightEntryParams struct {
	ID         int
	RecordedAt *time.Time
	Value      *float64
	Unit       *WeightUnit
	Note       *string
	UpdatedAt  time.Time
}

type DeleteWeightEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type FindManualWeightEntryParams struct {
	RecordedAt time.Time
	Unit       WeightUnit
}
