package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/qiita_export/models"
)

const (
	sleepTime = 100 * time.Millisecond
)

// リクエスト回数, QiitaAPIでは1時間あたり1000回がリミットのため、一応記録する
var RequestCount = 0

type QiitaAPI struct {
	requestBaseApiUrl string
	authHeaderToken   string
}

func NewQiitaAPI(domain, token string) *QiitaAPI {
	return &QiitaAPI{
		requestBaseApiUrl: fmt.Sprintf("https://%s/api/v2", domain),
		authHeaderToken:   fmt.Sprintf("Bearer %s", token),
	}
}

func (a QiitaAPI) newGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, a.wrapError(err)
	}
	req.Header.Set("Authorization", a.authHeaderToken)
	RequestCount++
	return req, nil
}

func (a QiitaAPI) wrapError(err error) error {
	return fmt.Errorf("合計リクエスト数: %d, エラー: %w", RequestCount, err)
}

// QiitaAPIを利用して、記事をAPI経由で取得する
// GET /api/v2/items にリクエストを送信し、格納する
// https://qiita.com/api/v2/docs#get-apiv2items
func (a QiitaAPI) RequestArticles(queryParams string) ([]models.Article, int, error) {
	// url.Valuesを使うと日本語がエンコーディングされてしまうため、そのままクエリパラメータを設定する
	requestUrl := fmt.Sprintf("%s/items?%s", a.requestBaseApiUrl, queryParams)

	req, err := a.newGetRequest(requestUrl)
	if err != nil {
		return nil, -1, a.wrapError(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, -1, a.wrapError(err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, -1, a.wrapError(fmt.Errorf("failed to get articles: %s, url: %s", res.Status, requestUrl))
	}

	total, err := strconv.Atoi(res.Header.Get("Total-Count"))
	if err != nil {
		return nil, -1, a.wrapError(err)
	}

	articles := make([]models.Article, 0, total)
	if err := json.NewDecoder(res.Body).Decode(&articles); err != nil {
		return nil, total, err
	}

	return articles, total, nil
}

// ArticleモデルのIDを利用して、絵文字リアクションをAPI経由で取得する
// GET /api/v2/items/:item_id/reactions にリクエストを送信し、格納する
// https://qiita.com/api/v2/docs#get-apiv2itemsitem_idreactions
func (a QiitaAPI) RequestArticleReactions(articleID string) ([]models.EmojiReaction, error) {
	requestUrl, err := url.JoinPath(a.requestBaseApiUrl, "items", articleID, "reactions")
	if err != nil {
		return nil, a.wrapError(err)
	}

	req, err := a.newGetRequest(requestUrl)
	if err != nil {
		return nil, a.wrapError(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, a.wrapError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, a.wrapError(fmt.Errorf("failed to get emoji reactions: %s", res.Status))
	}

	// 絵文字リアクション情報をEmojiReactionsに格納する
	reactions := make([]models.EmojiReaction, 0)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, a.wrapError(err)
	}
	if err := json.Unmarshal(body, &reactions); err != nil {
		return nil, a.wrapError(err)
	}

	return reactions, nil
}

// ArticleモデルのIDを利用して、コメントをAPI経由で取得する
// GET /api/v2/items/:item_id/comments にリクエストを送信し、格納する
// https://qiita.com/api/v2/docs#get-apiv2itemsitem_idcomments
func (a QiitaAPI) RequestComments(itemID string) ([]models.Comment, error) {
	requestUrl, err := url.JoinPath(a.requestBaseApiUrl, "items", itemID, "comments")
	if err != nil {
		return nil, a.wrapError(err)
	}

	req, err := a.newGetRequest(requestUrl)
	if err != nil {
		return nil, a.wrapError(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, a.wrapError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, a.wrapError(fmt.Errorf("failed to get comments: %s", res.Status))
	}

	// コメント情報をCommentsに格納する
	comments := make([]models.Comment, 0)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, a.wrapError(err)
	}
	if err := json.Unmarshal(body, &comments); err != nil {
		return nil, a.wrapError(err)
	}

	// コメントの絵文字リアクション情報を取得する
	for i, v := range comments {
		reactions, err := a.requestCommentReactions(v.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get emoji reactions: %w", err)
		}
		comments[i].EmojiReactions = reactions
	}

	return comments, nil
}

// コメントモデルのIDを利用して、絵文字リアクションをAPI経由で取得する
// GET /api/v2/comments/:comment_id/reactions にリクエストを送信し、格納する
// https://qiita.com/api/v2/docs#get-apiv2commentscomment_idreactions
func (a QiitaAPI) requestCommentReactions(commentID string) ([]models.EmojiReaction, error) {
	requestUrl, err := url.JoinPath(a.requestBaseApiUrl, "comments", commentID, "reactions")
	if err != nil {
		return nil, a.wrapError(err)
	}

	req, err := a.newGetRequest(requestUrl)
	if err != nil {
		return nil, a.wrapError(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, a.wrapError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, a.wrapError(fmt.Errorf("failed to get emoji reactions: %s", res.Status))
	}

	// 絵文字リアクション情報を格納する
	emojiReactions := make([]models.EmojiReaction, 0)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, a.wrapError(err)
	}
	if err := json.Unmarshal(body, &emojiReactions); err != nil {
		return nil, a.wrapError(err)
	}

	return emojiReactions, nil
}

// 添付ファイルのダウンロード
func (a QiitaAPI) DownloadArticleAssets(body, artDir string) (retErr error) {
	assetRegexp := regexp.MustCompile(os.Getenv("ASSET_REGEXP"))

	count := 0
	_ = assetRegexp.ReplaceAllStringFunc(body, func(s string) string {
		count++
		if retErr != nil {
			return s
		}

		f, err := os.Create(filepath.Join(artDir, path.Base(s)))
		if err != nil {
			retErr = err
			return s
		}

		req, err := a.newGetRequest(s)
		if err != nil {
			return s
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			retErr = err
			return s
		}
		defer res.Body.Close()

		if res.StatusCode == 403 {
			fmt.Println("403:", artDir, s)
		}

		if _, err := io.Copy(f, res.Body); err != nil {
			retErr = err
			return s
		}

		if err := f.Close(); err != nil {
			retErr = err
			return s
		}

		// レート制限で403になってしまうため待機時間を設ける
		time.Sleep(sleepTime)

		return s
	})
	fmt.Println("total assets", count)

	return
}
