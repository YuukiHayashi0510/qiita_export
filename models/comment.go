package models

import (
	"time"
)

// Qiita Teamのコメント
// https://qiita.com/api/v2/docs#%E3%82%B3%E3%83%A1%E3%83%B3%E3%83%88
type Comment struct {
	ID             string          `json:"id"`
	Body           string          `json:"body"`
	RenderedBody   string          `json:"rendered_body"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	User           User            `json:"user"`
	EmojiReactions []EmojiReaction `json:"reactions"` // 絵文字リアクション, 同一エンドポイントでは取得できない
}
