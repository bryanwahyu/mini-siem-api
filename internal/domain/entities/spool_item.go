package entities

import "time"

// SpoolItem stores metadata for spooled payloads destined for cold storage.
type SpoolItem struct {
    ID           string     `gorm:"primaryKey;size:32" json:"id"`
    CreatedAt    time.Time  `json:"created_at"`
    FilePath     string     `json:"file_path"`
    Kind         string     `json:"kind"`           // e.g., upload, rules-snapshot
    OriginalPath string     `json:"original_path"`  // intended object path in cold store
    ContentType  string     `json:"content_type"`
    Size         int64      `json:"size"`
    RetryCount   int        `json:"retry_count"`
    LastError    string     `json:"last_error"`
    UploadedAt   *time.Time `json:"uploaded_at"`
}

func (SpoolItem) TableName() string { return "spool_items" }

