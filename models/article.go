package models

import "time"

// Qiita Teamの記事
type Article struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	Body                string    `json:"body"`
	URL                 string    `json:"url"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Group               *Group    `json:"group"`
	Tags                []Tag     `json:"tags"`
	PageViewsCount      *int      `json:"page_views_count"`
	OrganizationURLName *string   `json:"organization_url_name"`
}
