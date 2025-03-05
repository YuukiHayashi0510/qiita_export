package models

import "time"

// Qiita Teamの絵文字リアクション
// https://qiita.com/api/v2/docs#%E7%B5%B5%E6%96%87%E5%AD%97%E3%83%AA%E3%82%A2%E3%82%AF%E3%82%B7%E3%83%A7%E3%83%B3
type EmojiReaction struct {
	CreatedAt time.Time `json:"created_at"`
	ImageUrl  *string   `json:"image_url"`
	Name      string    `json:"name"`
	User      User      `json:"user"`
}
