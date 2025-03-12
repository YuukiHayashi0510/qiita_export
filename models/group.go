package models

import "time"

// Qiita Teamのグループ
type Group struct {
	Name        string    `json:"name"`
	URLName     string    `json:"url_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Private     bool      `json:"private"`
}
