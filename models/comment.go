package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	EmojiReactions []EmojiReaction `json:"reactions"` // 絵文字リアクション, APIでは取得できない
}

// コメントモデルのIDを利用して、絵文字リアクションをAPI経由で取得する
// GET /api/v2/comments/:comment_id/reactions にリクエストを送信し、格納する
func (c *Comment) RequestEmojiReactions(domain, token string) ([]EmojiReaction, error) {
	requestUrl := fmt.Sprintf("https://%s/api/v2/comments/%s/reactions", domain, c.ID)
	authHeaderToken := fmt.Sprintf("Bearer %s", token)

	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeaderToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get emoji reactions: %s", res.Status)
	}

	// 絵文字リアクション情報を格納する
	emojiReactions := make([]EmojiReaction, 0)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &emojiReactions); err != nil {
		return nil, err
	}

	return emojiReactions, nil
}
