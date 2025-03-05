package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
	Comments            []Comment       `json:"comments"`  // コメント, APIでは取得できない
	EmojiReactions      []EmojiReaction `json:"reactions"` // 絵文字リアクション, APIでは取得できない
}

// ArticleモデルのIDを利用して、コメントをAPI経由で取得する
// GET /api/v2/items/:item_id/comments にリクエストを送信し、格納する
func (a Article) RequestComments(domain, token string) ([]Comment, error) {
	requestUrl := fmt.Sprintf("https://%s/api/v2/items/%s/comments", domain, a.ID)
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
		return nil, fmt.Errorf("failed to get comments: %s", res.Status)
	}

	// コメント情報をCommentsに格納する
	comments := make([]Comment, 0)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &comments); err != nil {
		return nil, err
	}

	// コメントの絵文字リアクション情報を取得する
	for i, v := range comments {
		time.Sleep(1 * time.Millisecond)
		reactions, err := v.RequestEmojiReactions(domain, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get emoji reactions: %w", err)
		}
		comments[i].EmojiReactions = reactions
	}

	return comments, nil
}

// ArticleモデルのIDを利用して、絵文字リアクションをAPI経由で取得する
// GET /api/v2/items/:item_id/reactions にリクエストを送信し、格納する
func (a Article) RequestEmojiReactions(domain, token string) ([]EmojiReaction, error) {
	requestUrl := fmt.Sprintf("https://%s/api/v2/items/%s/reactions", domain, a.ID)
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

	// 絵文字リアクション情報をEmojiReactionsに格納する
	reactions := make([]EmojiReaction, 0)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &reactions); err != nil {
		return nil, err
	}

	return reactions, nil
}
