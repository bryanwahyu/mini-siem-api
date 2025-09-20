package dto

import "time"

type DecisionDTO struct {
    ID        uint       `json:"id"`
    CreatedAt time.Time  `json:"created_at"`
    IP        string     `json:"ip"`
    Action    string     `json:"action"`
    Reason    string     `json:"reason"`
    Until     *time.Time `json:"until"`
}
