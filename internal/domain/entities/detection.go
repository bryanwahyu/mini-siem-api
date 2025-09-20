package entities

import "time"

type Detection struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    EventID   uint      `json:"event_id"`
    Category  string    `json:"category"`
    Rule      string    `json:"rule"`
    Severity  string    `json:"severity"`
    Metadata  string    `json:"metadata"` // JSON string
}

func (Detection) TableName() string { return "detections" }
