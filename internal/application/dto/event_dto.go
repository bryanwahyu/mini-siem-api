package dto

import "time"

type EventDTO struct {
    ID        uint      `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    Host      string    `json:"host"`
    Source    string    `json:"source"`
    IP        string    `json:"ip"`
    UA        string    `json:"ua"`
    Method    string    `json:"method"`
    Path      string    `json:"path"`
    Status    int       `json:"status"`
}
