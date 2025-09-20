package entities

import "time"

type Rule struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    Name      string    `json:"name"`
    Category  string    `json:"category"`
    Pattern   string    `json:"pattern"` // regex or keyword
    Enabled   bool      `json:"enabled"`
    Severity  string    `json:"severity"`
}

func (Rule) TableName() string { return "rules" }
