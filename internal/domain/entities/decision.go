package entities

import "time"

type Decision struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    IP        string    `json:"ip"`
    Action    string    `json:"action"` // block, unblock
    Reason    string    `json:"reason"`
    Until     *time.Time `json:"until"`
}

func (Decision) TableName() string { return "decisions" }
