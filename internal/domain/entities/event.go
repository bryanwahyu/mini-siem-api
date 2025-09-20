package entities

import "time"

type Event struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    Host      string    `json:"host"`
    Source    string    `json:"source"` // file path or journald unit
    Raw       string    `json:"raw"`    // raw log line
    IP        string    `json:"ip"`
    UA        string    `json:"ua"`
    Method    string    `json:"method"`
    Path      string    `json:"path"`
    Status    int       `json:"status"`
    Bytes     int64     `json:"bytes"`
    Referrer  string    `json:"referrer"`
}

func (Event) TableName() string { return "events" }
