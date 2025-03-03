package models

import "time"

// Qiita Teamの記事
type Article struct {
	ID                  string          `json:"id"`
	Title               string          `json:"title"`
	Body                string          `json:"body"`
	RenderedBody        string          `json:"rendered_body"`
	Coediting           bool            `json:"coediting"`
	CommentsCount       int             `json:"comments_count"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	Group               *Group          `json:"group"`
	LikesCount          int             `json:"likes_count"`
	Private             bool            `json:"private"`
	ReactionsCount      int             `json:"reactions_count"`
	StocksCount         int             `json:"stocks_count"`
	Tags                []Tag           `json:"tags"`
	URL                 string          `json:"url"`
	User                *User           `json:"user"`
	PageViewsCount      *int            `json:"page_views_count"`
	TeamMembership      *TeamMembership `json:"team_membership"`
	OrganizationURLName *string         `json:"organization_url_name"`
	Slide               bool            `json:"slide"`
}
