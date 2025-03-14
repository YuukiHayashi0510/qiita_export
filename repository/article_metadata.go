package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/qiita_export/models"
)

// ArticleMetadata はJSONメタデータファイルから記事情報を取得する実装
type ArticleMetadata struct{}

// GetArticleFromMetadata はメタデータファイルから記事情報を取得します
func (r *ArticleMetadata) GetArticle(metadataPath string) (*models.Article, error) {
	// メタデータファイルが存在するか確認
	if _, err := os.Stat(metadataPath); err != nil {
		return nil, fmt.Errorf("メタデータファイルがありません: %s", metadataPath)
	}

	// メタデータファイルを読み込む
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("メタデータファイル読み込みエラー: %w", err)
	}

	// JSONをパース
	var article models.Article
	if err := json.Unmarshal(metadataBytes, &article); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %w", err)
	}

	// IDが空の場合はエラー
	if article.ID == "" {
		return nil, fmt.Errorf("メタデータにIDフィールドがありません")
	}

	return &article, nil
}
